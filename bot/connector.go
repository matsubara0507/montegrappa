package bot

import (
	"fmt"
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
	Argv        []string
	Channel     string
	User        User
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

func (self *Event) Sayf(format string, a ...interface{}) {
	self.Bot.Sendf(self, format, a...)
}

func (self *Event) SayWithConfirm(text, reaction string, callback func(*Event)) {
	self.Bot.SendWithConfirm(self, text, reaction, callback)
}

func (self *Event) SayWithConfirmf(reaction string, callback func(*Event), format string, a ...interface{}) {
	self.Bot.SendWithConfirmf(self, reaction, callback, format, a...)
}

func (self *Event) SayRequireResponse(text string) (func(), chan string) {
	return self.Bot.SendRequireResponse(self, text)
}

func (self *Event) SayRequireResponsef(format string, a ...interface{}) (func(), chan string) {
	return self.Bot.SendRequireResponsef(self, format, a...)
}

func (self *Event) Reply(text string) {
	self.Bot.Sendf(self, "%l: %s", self.User, text)
}

func (self *Event) Replyf(format string, a ...interface{}) {
	self.Bot.Sendf(self, "%l:"+format, self.User, a...)
}

type SendedMessage struct {
	Message   string
	Channel   string
	Timestamp string
}

type User struct {
	Id   string
	Name string
}

func (user User) Format(f fmt.State, c rune) {
	if c == 'l' {
		fmt.Fprint(f, "<@"+user.Id+">")
		return
	}
	fmt.Fprint(f, user.Name)
}
