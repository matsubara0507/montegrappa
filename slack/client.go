package slack

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/matsubara0507/montegrappa/bot"
	"github.com/slack-go/slack"
)

type Client struct {
	domain string
	token  string
	client *slack.Client
}

func NewSlackClient(token string) *Client {
	return &Client{
		token:  token,
		client: slack.New(token),
	}
}

func (c *Client) Client() *slack.Client {
	return c.client
}

func (c *Client) Send(event *bot.Event, username string, text string) error {
	_, _, err := c.client.PostMessage(event.Channel, slack.MsgOptionUsername(username), slack.MsgOptionText(text, false))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SendWithConfirm(event *bot.Event, username, text string) (string, error) {
	_, ts, err := c.client.PostMessage(event.Channel, slack.MsgOptionUsername(username), slack.MsgOptionText(text, false))
	if err != nil {
		return "", err
	}

	return ts, nil
}

func (c *Client) SendPrivate(event *bot.Event, userId, text string) error {
	_, _, channelId, err := c.client.OpenIMChannel(userId)
	if err != nil {
		return err
	}

	_, _, err = c.client.PostMessage(channelId, slack.MsgOptionText(text, false))
	return err
}

func (c *Client) Attach(event *bot.Event, fileName string, file io.Reader, title string) error {
	_, err := c.client.UploadFile(slack.FileUploadParameters{
		Filename: fileName,
		Channels: []string{event.Channel},
		Reader:   file,
		Title:    title,
		Filetype: "auto",
	})

	return err
}

func (c *Client) WithIndicate(channel string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func(c string) {
		t := time.Tick(2 * time.Second)
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case <-t:
				//_ = c.sendTyping(c)
			}
		}
	}(channel)

	return cancel
}

func (c *Client) GetPermalink(event *bot.Event) string {
	return fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", c.teamDomain(), event.Channel, strings.Replace(event.Ts, ".", "", -1))
}

func (c *Client) teamDomain() string {
	if c.domain == "" {
		info, err := c.client.GetTeamInfo()
		if err != nil {
			return ""
		}
		c.domain = info.Domain
	}

	return c.domain
}

func (c *Client) GetChannelInfo(channelId string) (*bot.ChannelInfo, error) {
	channel, err := c.client.GetChannelInfo(channelId)
	if err != nil {
		return nil, err
	}

	var res bot.ChannelInfo
	res.Name = channel.Name
	res.Id = channel.ID
	return &res, nil
}
