package bot

import (
	"context"
	"fmt"
	"io"
)

type Connector interface {
	Connect() error
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
	GetPermalink(*Event) string
}

const (
	MessageEvent       = "message"
	UserTypingEvent    = "user_typing"
	ReactionAddedEvent = "reaction_added"
	ScheduledEvent     = "scheduled"
	UnknownEvent       = "unknown"
)

type SentMessage struct {
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
