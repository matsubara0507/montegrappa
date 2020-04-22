package bot

import (
	"context"
	"fmt"
	"io"
)

type Client interface {
	Send(*Event, string, string) error
	SendWithConfirm(*Event, string, string) (string, error)
	SendPrivate(*Event, string, string) error
	Attach(*Event, string, io.Reader, string) error
	WithIndicate(string) context.CancelFunc
	GetChannelInfo(string) (*ChannelInfo, error)
	GetPermalink(*Event) string
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
