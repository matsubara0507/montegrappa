package slack

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/f110/montegrappa/bot"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SlackConnector struct {
	ctx       context.Context
	cancel    context.CancelFunc
	eventChan chan *bot.Event
	idle      chan bool

	token      string
	bufChan    chan []byte
	startTime  int
	connection *websocket.Conn
}

type Ping struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Time int    `json:"time"`
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

var (
	ReadBufferSize       = 512
	ReadTimeout          = 1 * time.Minute
	WriteTimeout         = 1 * time.Minute
	HeartbeatInterval    = 30 * time.Second
	ErrFailedPostMessage = errors.New("Failed chat.postMessage")
)

func NewSlackConnector(token string) *SlackConnector {
	startTime := int(time.Now().Unix())
	ctx, cancel := context.WithCancel(context.Background())

	return &SlackConnector{
		ctx:       ctx,
		cancel:    cancel,
		token:     token,
		startTime: startTime,
		eventChan: make(chan *bot.Event),
	}
}

func (this *SlackConnector) Connect() {
	v := url.Values{}
	v.Set("token", this.token)
	response, _ := http.PostForm("https://slack.com/api/rtm.start", v)
	dec := json.NewDecoder(response.Body)
	var data struct {
		URL string `json:"url"`
	}
	dec.Decode(&data)
	log.Print("start connect to ", data.URL)
	ws, err := websocket.Dial(data.URL, "", "http://localhost")
	if err != nil {
		log.Print(err)
	}

	this.connection = ws
	this.startReading()
}

func (this *SlackConnector) Async() bool {
	return true
}

func (this *SlackConnector) Listen() error {
	for {
		select {
		case buf := <-this.bufChan:
			var event Event
			json.Unmarshal(buf, &event)

			if event.Ts != "" {
				ts := strings.Split(event.Ts, ".")[0]
				i, _ := strconv.Atoi(ts)
				if i < this.startTime {
					log.Print("skip event")
					continue
				}
			}

			botEvent := new(bot.Event)
			switch event.Type {
			case "message":
				var messageEvent Message
				json.Unmarshal(buf, &messageEvent)
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
				json.Unmarshal(buf, &userTypingEvent)
				if userTypingEvent.User == "" {
					continue
				}

				botEvent.Type = bot.UserTypingEvent
				botEvent.Channel = userTypingEvent.Channel
				botEvent.User.Id = userTypingEvent.User
			case "pong":
				log.Print("receive pong")
				continue
			case "reaction_added":
				botEvent.Type = bot.ReactionAddedEvent
				reactionAdded := new(ReactionAdded)
				json.Unmarshal(buf, reactionAdded)
				if reactionAdded.Item.Type != "message" {
					continue
				}
				botEvent.Channel = reactionAdded.Item.Channel
				botEvent.Ts = reactionAdded.Item.Ts
				botEvent.User.Id = reactionAdded.ItemUser
				botEvent.Reaction = reactionAdded.Reaction
			default:
				botEvent.Type = bot.UnknownEvent
			}

			this.eventChan <- botEvent
		case <-this.ctx.Done():
			log.Print("disconnect server")
			return errors.New("disconnect")
		}
	}
}

func (this *SlackConnector) ReceivedEvent() chan *bot.Event {
	return this.eventChan
}

func (this *SlackConnector) Idle() chan bool {
	return this.idle
}

func (this *SlackConnector) Send(event *bot.Event, username string, text string) error {
	v := url.Values{}
	v.Set("token", this.token)
	v.Set("channel", event.Channel)
	v.Set("text", text)
	v.Set("as_user", "false")

	response, err := http.PostForm("https://slack.com/api/chat.postMessage", v)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	dec := json.NewDecoder(response.Body)
	var data struct {
		Ok bool   `json:"ok"`
		Ts string `json:"ts"`
	}
	dec.Decode(&data)

	if data.Ok != true {
		return ErrFailedPostMessage
	}

	return nil
}

func (this *SlackConnector) SendWithConfirm(event *bot.Event, username, text string) (string, error) {
	v := url.Values{}
	v.Set("token", this.token)
	v.Set("channel", event.Channel)
	v.Set("text", text)
	v.Set("as_user", "false")
	v.Set("username", username)

	response, _ := http.PostForm("https://slack.com/api/chat.postMessage", v)
	dec := json.NewDecoder(response.Body)
	var data struct {
		Ok bool   `json:"ok"`
		Ts string `json:"ts"`
	}
	dec.Decode(&data)

	if data.Ok != true {
		return "", ErrFailedPostMessage
	}

	return data.Ts, nil
}

func (this *SlackConnector) startReading() {
	log.Print("start reading")
	var msg []byte
	this.bufChan = make(chan []byte)

	go func() {
	READ:
		for {
			var tmp = make([]byte, ReadBufferSize)
			this.connection.SetReadDeadline(time.Now().Add(ReadTimeout))
			n, err := this.connection.Read(tmp)
			if err == io.EOF {
				this.cancel()
				break READ
			}
			if err != nil {
				log.Fatal(err)
			}
			if msg != nil {
				msg = append(msg, tmp[:n]...)
			} else {
				msg = tmp[:n]
			}
			if n != ReadBufferSize {
				this.bufChan <- msg
				msg = nil
			}
		}
	}()

	go func() {
		time.Sleep(10 * time.Second)
		this.heartbeat()
	}()
}

func (this *SlackConnector) heartbeat() {
	id := 0
	c := time.Tick(HeartbeatInterval)
HEARTBEAT:
	for {
		var ret error
		select {
		case <-this.ctx.Done():
			break HEARTBEAT
		case <-c:
			ret = this.sendPing(id)
		}

		if ret != nil {
			break
		}

		id++
	}
}

func (this *SlackConnector) sendPing(id int) error {
	ping := &Ping{Id: id, Type: "ping", Time: int(time.Now().Unix())}
	buf, err := json.Marshal(ping)
	if err != nil {
		log.Print("failed json.Marshal")
	}
	log.Print("send ping to slack")
	this.connection.SetWriteDeadline(time.Now().Add(WriteTimeout))
	_, err = this.connection.Write(buf)
	if err != nil {
		log.Print("failed send ping")
		this.cancel()
		return err
	}

	return nil
}
