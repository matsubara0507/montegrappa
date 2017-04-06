package bot

import (
	"context"
	"fmt"
	"io"
	"time"
)

type Connector interface {
	Connect()
	Listen() error
	ReceivedEvent() chan *Event
	Send(*Event, string, string) error
	SendWithConfirm(*Event, string, string) (string, error)
	Attach(*Event, string, io.Reader, string) error
	WithIndicate(string) context.CancelFunc
	SendPrivate(*Event, string, string) error
	Async() bool
	Idle() chan bool
	GetChannelInfo(string) (*ChannelInfo, error)
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

func (event *Event) ChannelName() (string, error) {
	channelInfo, err := event.Bot.Connector.GetChannelInfo(event.Channel)
	if err != nil {
		return "", err
	}

	return channelInfo.Name, nil
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
	self.Bot.Sendf(self, fmt.Sprintf("%l: %s", self.User, format), a...)
}

func (self *Event) WithIndicate(f func() error) {
	self.Bot.WithIndicate(self.Channel, f)
}

func (self *Event) Attach(title, fileName string, file io.Reader) error {
	return self.Bot.Attach(self, title, fileName, file)
}

func (self *Event) Direct(text string) {
	self.Bot.SendPrivate(self, text)
}

func (self *Event) Directf(format string, a ...interface{}) {
	self.Bot.SendPrivate(self, fmt.Sprintf(format, a...))
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

type ChannelInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
