package bot

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

type EventHandler struct {
	OnError OnError

	accept      bool
	acceptUsers map[string]bool
	ignoreUsers map[string]bool
	commands    map[string][]Command
	mutex       *sync.RWMutex
}

type Command struct {
	CommandType string

	eventType   string
	description string
	pattern     *regexp.Regexp
	channel     string
	user        string
	argv        bool
	messageId   string
	reaction    string
	callback    func(*Event)
	createdAt   time.Time
}

const (
	CommandTypeRequireResponse = "require_response"
)

var (
	ReactionExpire = 3 * time.Minute
)

func NewEventHandler(ignoreUsers []string, acceptUsers []string) *EventHandler {
	accept := false
	acceptMap := make(map[string]bool)
	ignoreMap := make(map[string]bool)
	if len(acceptUsers) > 0 {
		accept = true
		for _, v := range acceptUsers {
			acceptMap[v] = true
		}
	} else {
		for _, v := range ignoreUsers {
			ignoreMap[v] = true
		}
	}

	return &EventHandler{
		accept:      accept,
		acceptUsers: acceptMap,
		ignoreUsers: ignoreMap,
		commands:    make(map[string][]Command, 0),
		mutex:       &sync.RWMutex{},
	}
}

func (eventHandler *EventHandler) AddCommand(pattern *regexp.Regexp, description string, callback func(*Event), argv bool) {
	command := &Command{pattern: pattern, description: description, callback: callback, argv: argv}
	eventHandler.AddHandler(MessageEvent, command)
}

func (eventHandler *EventHandler) Appearance(user string, callback func(*Event)) {
	command := &Command{user: user, callback: callback}
	eventHandler.AddHandler(UserTypingEvent, command)
}

func (eventHandler *EventHandler) WatchReaction(reaction string, callback func(*Event)) {
	command := &Command{reaction: reaction, callback: callback}
	eventHandler.AddHandler(ReactionAddedEvent, command)
}

func (eventHandler *EventHandler) RequireReaction(channel, id, reaction, userId string, callback func(*Event)) {
	c := &Command{messageId: channel + id, reaction: reaction, user: userId, callback: callback, createdAt: time.Now()}
	go eventHandler.AddHandler(ReactionAddedEvent, c)
}

func (eventHandler *EventHandler) RemoveRequireReaction(eventId, reaction string) {
	eventHandler.mutex.Lock()
	defer eventHandler.mutex.Unlock()
	newCommands := make([]Command, 0)
	for _, c := range eventHandler.commands[ReactionAddedEvent] {
		if c.messageId == eventId && c.reaction == reaction {
			continue
		}
		newCommands = append(newCommands, c)
	}
	eventHandler.commands[ReactionAddedEvent] = newCommands
}

func (eventHandler *EventHandler) RequireResponse(channel, user string) (func(), chan string) {
	resChan := make(chan string)
	callback := func(msg *Event) {
		resChan <- msg.Message
	}
	cancelFunc := func() {
		go eventHandler.RemoveRequireResponse(channel, user)
	}
	c := &Command{CommandType: CommandTypeRequireResponse, channel: channel, user: user, callback: callback}
	go eventHandler.AddHandler(MessageEvent, c)
	return cancelFunc, resChan
}

func (eventHandler *EventHandler) RemoveRequireResponse(channel, user string) {
	eventHandler.mutex.Lock()
	defer eventHandler.mutex.Unlock()

	newCommands := make([]Command, 0)
	for _, c := range eventHandler.commands[MessageEvent] {
		if c.CommandType == CommandTypeRequireResponse && c.channel == channel && c.user == user {
			continue
		}
		newCommands = append(newCommands, c)
	}
	eventHandler.commands[MessageEvent] = newCommands
}

func (eventHandler *EventHandler) AddHandler(eventType string, command *Command) {
	eventHandler.mutex.Lock()
	defer eventHandler.mutex.Unlock()
	if eventHandler.commands[eventType] == nil {
		eventHandler.commands[eventType] = make([]Command, 0)
	}

	eventHandler.commands[eventType] = append(eventHandler.commands[eventType], *command)
}

func (eventHandler *EventHandler) Handle(event *Event, async bool) {
	if _, ok := eventHandler.acceptUsers[event.User.Id]; eventHandler.accept && ok == false {
		return
	}
	if _, ok := eventHandler.ignoreUsers[event.User.Id]; ok == true {
		return
	}

	eventHandler.mutex.RLock()
	defer eventHandler.mutex.RUnlock()
	for _, command := range eventHandler.commands[event.Type] {
		switch event.Type {
		case MessageEvent:
			if command.CommandType == CommandTypeRequireResponse && event.Channel == command.channel && event.User.Id == command.user {
				eventHandler.commandCallback(command, event, async)
				return
			}
			if command.CommandType == CommandTypeRequireResponse {
				return
			}

			if command.pattern.MatchString(event.Message) == true {
				if command.argv == true {
					matched := command.pattern.FindStringSubmatch(event.Message)
					event.Argv = strings.Fields(matched[1])
				}

				eventHandler.commandCallback(command, event, async)
				return
			}
		case UserTypingEvent:
			if event.User.Id == command.user {
				eventHandler.commandCallback(command, event, async)
			}
		case ReactionAddedEvent:
			if !command.createdAt.IsZero() && time.Now().Sub(command.createdAt) >= ReactionExpire {
				go eventHandler.RemoveRequireReaction(command.messageId, command.reaction)
				continue
			}
			if event.EventId() == command.messageId && event.User.Id == command.user && event.Reaction == command.reaction {
				go eventHandler.RemoveRequireReaction(event.EventId(), event.Reaction)
				eventHandler.commandCallback(command, event, async)
				return
			}
			if event.Reaction == command.reaction {
				eventHandler.commandCallback(command, event, async)
			}
		}
	}
}

func (eventHandler *EventHandler) commandCallback(command Command, event *Event, async bool) {
	if async {
		go func(command Command, event *Event, onError OnError) {
			eventHandler.commandCallbackWithLog(command, event, onError)
		}(command, event, eventHandler.OnError)
	} else {
		eventHandler.commandCallbackWithLog(command, event, eventHandler.OnError)
	}
}

func (eventHandler *EventHandler) commandCallbackWithLog(command Command, event *Event, onError OnError) {
	logging := true
	defer func() {
		if logging && onError != nil {
			onError(event)
		}
	}()
	command.callback(event)
	logging = false
}
