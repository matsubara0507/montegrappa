package bot

import (
	"fmt"
	"io"
	"log"
	"regexp"
)

var (
	botInstance *Bot
)

type Bot struct {
	Connector        Connector
	Name             string
	TeamID           string
	connectErrorChan chan bool
	eventHandler     *EventHandler
}

func NewBot(connector Connector, name, teamId string, ignoreUsers []string) *Bot {
	bot := &Bot{
		Connector:        connector,
		Name:             name,
		TeamID:           teamId,
		connectErrorChan: make(chan bool),
		eventHandler:     NewEventHandler(ignoreUsers),
	}
	botInstance = bot

	return bot
}

func (self *Bot) Connect() {
	self.Connector.Connect()

	go func() {
		res := self.Connector.Listen()
		if res != nil {
			self.connectErrorChan <- true
		}
	}()
}

func (self *Bot) Start() {
	self.Connect()
	for {
		select {
		case event := <-self.Connector.ReceivedEvent():
			event.Bot = self
			if self.Connector.Async() == true {
				go self.eventHandler.Handle(event, true)
			} else {
				self.eventHandler.Handle(event, false)
				self.Connector.Idle() <- true
			}
		case <-self.connectErrorChan:
			log.Print("reconnect")
			self.Connect()
		}
	}
}

func (self *Bot) Send(event *Event, text string) {
	self.Connector.Send(event, self.Name, text)
}

func (self *Bot) Sendf(event *Event, format string, a ...interface{}) {
	text := fmt.Sprintf(format, a...)
	self.Send(event, text)
}

func (self *Bot) SendWithConfirm(event *Event, text, reaction string, callback func(*Event)) {
	id, _ := self.Connector.SendWithConfirm(event, self.Name, text)
	self.eventHandler.RequireReaction(event.Channel, id, reaction, callback)
}

func (self *Bot) SendWithConfirmf(event *Event, reaction string, callback func(*Event), format string, a ...interface{}) {
	text := fmt.Sprintf(format, a)
	self.SendWithConfirm(event, text, reaction, callback)
}

func (self *Bot) SendRequireResponse(event *Event, text string) (func(), chan string) {
	self.Connector.Send(event, self.Name, text)
	return self.eventHandler.RequireResponse(event.Channel, event.User.Id)
}

func (self *Bot) SendRequireResponsef(event *Event, format string, a ...interface{}) (func(), chan string) {
	text := fmt.Sprintf(format, a)
	return self.SendRequireResponse(event, text)
}

func (self *Bot) WithIndicate(channel string, f func() error) {
	cancel := self.Connector.WithIndicate(channel)
	defer cancel()
	f()
}

func (self *Bot) Attach(event *Event, title, fileName string, file io.Reader) error {
	return self.Connector.Attach(event, fileName, file, title)
}

func (self *Bot) SendPrivate(event *Event, text string) {
	self.Connector.SendPrivate(event, event.User.Id, text)
}

func (self *Bot) Hear(pattern string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile(pattern), "", callback, false)
}

func (self *Bot) Command(pattern string, description string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile("\\A"+self.Name+"\\s+"+pattern+"\\z"), pattern+" - "+description, callback, false)
}

func (self *Bot) CommandWithArgv(pattern string, description string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile("\\A"+self.Name+"\\s+"+pattern+"(?:\\s+(.+))*\\z"), pattern+" - "+description, callback, true)
}

func (self *Bot) Appearance(user string, callback func(*Event)) {
	self.eventHandler.Appearance(user, callback)
}
