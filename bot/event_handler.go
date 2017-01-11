package bot

import (
	"regexp"
	"sync"
	"time"
)

type EventHandler struct {
	ignoreUsers map[string]bool
	commands    map[string][]Command
	mutex       *sync.RWMutex
}

type Command struct {
	eventType   string
	description string
	pattern     *regexp.Regexp
	user        string
	argv        bool
	messageId   string
	reaction    string
	callback    func(*Event)
	createdAt   time.Time
}

var (
	ReactionExpire = 3 * time.Minute
)

func NewEventHandler(ignoreUsers []string) *EventHandler {
	ignore := make(map[string]bool)
	for _, v := range ignoreUsers {
		ignore[v] = true
	}

	return &EventHandler{
		ignoreUsers: ignore,
		commands:    make(map[string][]Command, 0),
		mutex:       &sync.RWMutex{},
	}
}

func (this *EventHandler) AddCommand(pattern *regexp.Regexp, description string, callback func(*Event), argv bool) {
	command := &Command{pattern: pattern, description: description, callback: callback, argv: argv}
	this.AddHandler(MessageEvent, command)
}

func (this *EventHandler) Appearance(user string, callback func(*Event)) {
	command := &Command{user: user, callback: callback}
	this.AddHandler(UserTypingEvent, command)
}

func (this *EventHandler) RequireReaction(channel, id, reaction string, callback func(*Event)) {
	c := &Command{messageId: channel + id, reaction: reaction, callback: callback, createdAt: time.Now()}
	go this.AddHandler(ReactionAddedEvent, c)
}

func (this *EventHandler) RemoveRequireReaction(eventId, reaction string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	newCommands := make([]Command, 0)
	for _, c := range this.commands[ReactionAddedEvent] {
		if c.messageId == eventId && c.reaction == reaction {
			continue
		}
		newCommands = append(newCommands, c)
	}
	this.commands[ReactionAddedEvent] = newCommands
}

func (this *EventHandler) AddHandler(eventType string, command *Command) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.commands[eventType] == nil {
		this.commands[eventType] = make([]Command, 0)
	}

	this.commands[eventType] = append(this.commands[eventType], *command)
}

func (this *EventHandler) Handle(event *Event) {
	if _, ok := this.ignoreUsers[event.User]; ok == true {
		return
	}

	this.mutex.RLock()
	defer this.mutex.RUnlock()
	for _, command := range this.commands[event.Type] {
		switch event.Type {
		case MessageEvent:
			if command.pattern.MatchString(event.Message) == true {
				if command.argv == true {
					matched := command.pattern.FindStringSubmatch(event.Message)
					event.Argv = matched[1]
					command.callback(event)
				} else {
					command.callback(event)
				}
				return
			}
		case UserTypingEvent:
			if event.User == command.user {
				command.callback(event)
			}
		case ReactionAddedEvent:
			if time.Now().Sub(command.createdAt) >= ReactionExpire {
				go this.RemoveRequireReaction(command.messageId, command.reaction)
				continue
			}
			if event.EventId() == command.messageId && event.Reaction == command.reaction {
				command.callback(event)
				go this.RemoveRequireReaction(event.EventId(), event.Reaction)
			}
		}
	}
}
