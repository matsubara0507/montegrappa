package bot

import (
	"time"
)

type Connector interface {
	Connect()
	Listen() error
	ReceivedEvent() chan *Event
	Send(*Event, string, string) error
	SendWithConfirm(*Event, string, string) (string, error)
	Async() bool
	Idle() chan bool
}

const (
	MessageEvent       = "message"
	UserTypingEvent    = "user_typing"
	ReactionAddedEvent = "reaction_added"
	UnknownEvent       = "unknown"
)

type Event struct {
	Type        string
	Message     string
	Argv        string
	Channel     string
	User        string
	Reaction    string
	Ts          string
	Timestamp   time.Time
	MentionName string
	Bot         *Bot
}

func (event *Event) EventId() string {
	return event.Channel + event.Ts
}

func (self *Event) Say(text string) {
	self.Bot.Send(self, text)
}

func (self *Event) SayWithConfirm(text, reaction string, callback func(*Event)) {
	self.Bot.SendWithConfirm(self, text, reaction, callback)
}

func (self *Event) SayRequireResponse(text string) (func(), chan string) {
	return self.Bot.SendRequireResponse(self, text)
}

type SendedMessage struct {
	Message   string
	Channel   string
	Timestamp string
}
