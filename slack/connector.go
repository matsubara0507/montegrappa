package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matsubara0507/montegrappa/bot"
	"github.com/slack-go/slack"
	"golang.org/x/net/websocket"
)

var (
	ReadBufferSize    = 4096
	ReadTimeout       = 1 * time.Minute
	WriteTimeout      = 1 * time.Minute
	HeartbeatInterval = 30 * time.Second
)

type Connector struct {
	mutex          *sync.Mutex
	disconnectConn bool
	eventChan      chan *bot.Event
	idle           chan bool

	token      string
	teamId     string
	domain     string
	bufChan    chan []byte
	errorChan  chan error
	startTime  int
	connection *websocket.Conn
	client     *slack.Client
}

type Ping struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Time int    `json:"time"`
}

type Typing struct {
	Id      int    `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

type Event struct {
	Type string
	Ts   string
	Raw  []byte
}

type Message struct {
	Type    string
	SubType string `json:"subtype"`
	Ts      string
	Channel string
	User    string
	Text    string
	ts      string
}

type UserTyping struct {
	Type    string
	Channel string
	User    string
}

type ReactionAdded struct {
	Type     string `json:"type"`
	User     string `json:"user"`
	Reaction string `json:"reaction"`
	ItemUser string `json:"item_user"`
	Item     struct {
		Type        string `json:"type"`
		Channel     string `json:"channel"`
		Ts          string `json:"ts"`
		File        string `json:"file"`
		FileComment string `json:"file_comment"`
	} `json:"item"`
	EventTs string `json:"event_ts"`
}

func NewConnector(teamId, token string) *Connector {
	startTime := int(time.Now().Unix())

	return &Connector{
		token:     token,
		teamId:    teamId,
		startTime: startTime,
		eventChan: make(chan *bot.Event),
		errorChan: make(chan error),
		mutex:     &sync.Mutex{},
		client:    slack.New(token),
	}
}

func (connector *Connector) Setup() error {
	_, u, err := connector.client.ConnectRTM()
	if err != nil {
		return err
	}
	log.Printf("start connect to %s", u)
	ws, err := websocket.Dial(u, "", "http://localhost")
	if err != nil {
		return err
	}

	connector.disconnectConn = false
	connector.connection = ws
	connector.startReading()

	return nil
}

func (*Connector) Async() bool {
	return true
}

func (connector *Connector) Client() *slack.Client {
	return connector.client
}

func (connector *Connector) Start() error {
	for {
		select {
		case buf := <-connector.bufChan:
			var event Event
			if err := json.Unmarshal(buf, &event); err != nil {
				continue
			}

			if event.Ts != "" {
				ts := strings.Split(event.Ts, ".")[0]
				i, _ := strconv.Atoi(ts)
				if i < connector.startTime {
					continue
				}
			}

			botEvent := new(bot.Event)
			switch event.Type {
			case "message":
				var messageEvent Message
				if err := json.Unmarshal(buf, &messageEvent); err != nil {
					continue
				}
				if messageEvent.User == "" {
					continue
				}

				botEvent.Type = bot.MessageEvent
				botEvent.Message = messageEvent.Text
				botEvent.Channel = messageEvent.Channel
				botEvent.User.Id = messageEvent.User
				botEvent.Ts = messageEvent.Ts
			case "user_typing":
				var userTypingEvent UserTyping
				if err := json.Unmarshal(buf, &userTypingEvent); err != nil {
					continue
				}
				if userTypingEvent.User == "" {
					continue
				}

				botEvent.Type = bot.UserTypingEvent
				botEvent.Channel = userTypingEvent.Channel
				botEvent.User.Id = userTypingEvent.User
			case "pong":
				continue
			case "reaction_added":
				botEvent.Type = bot.ReactionAddedEvent
				reactionAdded := new(ReactionAdded)
				if err := json.Unmarshal(buf, reactionAdded); err != nil {
					continue
				}
				if reactionAdded.Item.Type != "message" {
					continue
				}
				botEvent.Channel = reactionAdded.Item.Channel
				botEvent.Ts = reactionAdded.Item.Ts
				botEvent.User.Id = reactionAdded.User
				botEvent.Reaction = reactionAdded.Reaction
			default:
				botEvent.Type = bot.UnknownEvent
			}

			connector.eventChan <- botEvent
		case <-connector.errorChan:
			return errors.New("disconnect")
		}
	}
}

func (connector *Connector) ReceivedEvent() chan *bot.Event {
	return connector.eventChan
}

func (connector *Connector) Idle() chan bool {
	return connector.idle
}

func (connector *Connector) Send(event *bot.Event, username string, text string) error {
	_, _, err := connector.client.PostMessage(event.Channel, slack.MsgOptionUsername(username), slack.MsgOptionText(text, false))
	if err != nil {
		return err
	}

	return nil
}

func (connector *Connector) SendWithConfirm(event *bot.Event, username, text string) (string, error) {
	_, ts, err := connector.client.PostMessage(event.Channel, slack.MsgOptionUsername(username), slack.MsgOptionText(text, false))
	if err != nil {
		return "", err
	}

	return ts, nil
}

func (connector *Connector) SendPrivate(event *bot.Event, userId, text string) error {
	_, _, channelId, err := connector.client.OpenIMChannel(userId)
	if err != nil {
		return err
	}

	_, _, err = connector.client.PostMessage(channelId, slack.MsgOptionText(text, false))
	return err
}

func (connector *Connector) Attach(event *bot.Event, fileName string, file io.Reader, title string) error {
	_, err := connector.client.UploadFile(slack.FileUploadParameters{
		Filename: fileName,
		Channels: []string{event.Channel},
		Reader:   file,
		Title:    title,
		Filetype: "auto",
	})

	return err
}

func (connector *Connector) WithIndicate(channel string) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func(c string) {
		t := time.Tick(2 * time.Second)
	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case <-t:
				_ = connector.sendTyping(c)
			}
		}
	}(channel)

	return cancel
}

func (connector *Connector) GetPermalink(event *bot.Event) string {
	return fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", connector.teamDomain(), event.Channel, strings.Replace(event.Ts, ".", "", -1))
}

func (connector *Connector) teamDomain() string {
	if connector.domain == "" {
		info, err := connector.client.GetTeamInfo()
		if err != nil {
			return ""
		}
		connector.domain = info.Domain
	}

	return connector.domain
}

func (connector *Connector) GetChannelInfo(channelId string) (*bot.ChannelInfo, error) {
	channel, err := connector.client.GetChannelInfo(channelId)
	if err != nil {
		return nil, err
	}

	var res bot.ChannelInfo
	res.Name = channel.Name
	res.Id = channel.ID
	return &res, nil
}

func (connector *Connector) startReading() {
	log.Print("start reading")
	var msg []byte
	connector.bufChan = make(chan []byte)

	go func() {
		tmp := make([]byte, ReadBufferSize)
		for {
			if err := connector.connection.SetReadDeadline(time.Now().Add(ReadTimeout)); err != nil {
				break
			}
			n, err := connector.connection.Read(tmp)
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			if msg != nil {
				msg = append(msg, tmp[:n]...)
			} else {
				msg = make([]byte, n)
				copy(msg, tmp[:n])
			}
			if n != ReadBufferSize {
				connector.bufChan <- msg
				msg = nil
			}
		}
		connector.disconnectConnection()
	}()

	go func() {
		time.Sleep(10 * time.Second)
		connector.heartbeat()
		connector.disconnectConnection()
	}()
}

func (connector *Connector) disconnectConnection() {
	connector.mutex.Lock()
	defer connector.mutex.Unlock()

	if connector.disconnectConn == false {
		connector.disconnectConn = true
		connector.errorChan <- errors.New("disconnecting")
	}
}

func (connector *Connector) heartbeat() {
	id := 0
	for {
		err := connector.sendPing(id)
		if err != nil {
			break
		}

		id++
		time.Sleep(HeartbeatInterval)
	}
}

func (connector *Connector) sendPing(id int) error {
	ping := &Ping{Id: id, Type: "ping", Time: int(time.Now().Unix())}
	buf, err := json.Marshal(ping)
	if err != nil {
		return err
	}

	if err := connector.connection.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	_, err = connector.connection.Write(buf)

	return err
}

func (connector *Connector) sendTyping(channel string) error {
	typing := &Typing{Id: 1, Type: "typing", Channel: channel}
	buf, err := json.Marshal(typing)
	if err != nil {
		return err
	}

	if err := connector.connection.SetWriteDeadline(time.Now().Add(WriteTimeout)); err != nil {
		return err
	}
	_, err = connector.connection.Write(buf)

	return err
}
