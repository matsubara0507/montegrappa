package bot

import (
	"log"
	"regexp"
)

var (
	botInstance *Bot
)

type Bot struct {
	connector        Connector
	name             string
	connectErrorChan chan bool
	eventHandler     *EventHandler
}

func BotInstance() *Bot {
	return botInstance
}

func NewBot(connector Connector, name string, ignoreUsers []string) *Bot {
	if botInstance != nil {
		return botInstance
	}

	bot := &Bot{
		connector:        connector,
		name:             name,
		connectErrorChan: make(chan bool),
		eventHandler:     NewEventHandler(ignoreUsers),
	}
	botInstance = bot

	return bot
}

func (self *Bot) Connect() {
	self.connector.Connect()

	go func() {
		res := self.connector.Listen()
		if res != nil {
			self.connectErrorChan <- true
		}
	}()
}

func (self *Bot) Start() {
	self.Connect()
	for {
		select {
		case event := <-self.connector.ReceivedEvent():
			event.Bot = self
			if self.connector.Async() == true {
				go self.eventHandler.Handle(event)
			} else {
				self.eventHandler.Handle(event)
				self.connector.Idle() <- true
			}
		case <-self.connectErrorChan:
			log.Print("reconnect")
			self.Connect()
		}
	}
}

func (self *Bot) Send(event *Event, text string) {
	self.connector.Send(event, self.name, text)
}

func (self *Bot) SendWithConfirm(event *Event, text, reaction string, callback func(*Event)) {
	id, _ := self.connector.SendWithConfirm(event, self.name, text)
	self.eventHandler.RequireReaction(event.Channel, id, reaction, callback)
}

func (self *Bot) Hear(pattern string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile(pattern), "", callback, false)
}

func (self *Bot) Command(pattern string, description string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile("\\A"+self.name+"\\s+"+pattern+"\\z"), pattern+" - "+description, callback, false)
}

func (self *Bot) CommandWithArgv(pattern string, description string, callback func(*Event)) {
	self.eventHandler.AddCommand(regexp.MustCompile("\\A"+self.name+"\\s+"+pattern+"(?:\\s+(.+))*\\z"), pattern+" - "+description, callback, true)
}

func (self *Bot) Appearance(user string, callback func(*Event)) {
	self.eventHandler.Appearance(user, callback)
}
