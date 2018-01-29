package bot

import (
	"fmt"
	"io"
	"time"
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

func (event *Event) Say(text string) {
	event.Bot.Send(event, text)
}

func (event *Event) Sayf(format string, a ...interface{}) {
	event.Bot.Sendf(event, format, a...)
}

func (event *Event) SayWithConfirm(text, reaction string, callback func(*Event)) {
	event.Bot.SendWithConfirm(event, text, reaction, callback)
}

func (event *Event) SayWithConfirmf(reaction string, callback func(*Event), format string, a ...interface{}) {
	event.Bot.SendWithConfirmf(event, reaction, callback, format, a...)
}

func (event *Event) SayRequireResponse(text string) (func(), chan string) {
	return event.Bot.SendRequireResponse(event, text)
}

func (event *Event) SayRequireResponsef(format string, a ...interface{}) (func(), chan string) {
	return event.Bot.SendRequireResponsef(event, format, a...)
}

func (event *Event) Reply(text string) {
	event.Bot.Sendf(event, "%l: %s", event.User, text)
}

func (event *Event) Replyf(format string, a ...interface{}) {
	event.Bot.Sendf(event, fmt.Sprintf("%l: %s", event.User, format), a...)
}

func (event *Event) WithIndicate(f func() error) {
	event.Bot.WithIndicate(event.Channel, f)
}

func (event *Event) Attach(title, fileName string, file io.Reader) error {
	return event.Bot.Attach(event, title, fileName, file)
}

func (event *Event) Direct(text string) {
	event.Bot.SendPrivate(event, text)
}

func (event *Event) Directf(format string, a ...interface{}) {
	event.Bot.SendPrivate(event, fmt.Sprintf(format, a...))
}

func (event *Event) Permalink() string {
	return event.Bot.GetPermalink(event)
}
