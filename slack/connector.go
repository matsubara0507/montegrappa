package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/f110/montegrappa/bot"
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
		startTime: startTime,
		eventChan: make(chan *bot.Event),
		errorChan: make(chan error),
		mutex:     &sync.Mutex{},
	}
}

func (connector *Connector) Connect() error {
	url, err := connector.RTMConnect()
	if err != nil {
		return err
	}
	log.Printf("start connect to %s", url)
	ws, err := websocket.Dial(url, "", "http://localhost")
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

func (connector *Connector) Listen() error {
	for {
		select {
		case buf := <-connector.bufChan:
			var event Event
			json.Unmarshal(buf, &event)

			if event.Ts != "" {
				ts := strings.Split(event.Ts, ".")[0]
				i, _ := strconv.Atoi(ts)
				if i < connector.startTime {
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
				botEvent.User.Id = reactionAdded.User
				botEvent.Reaction = reactionAdded.Reaction
			default:
				botEvent.Type = bot.UnknownEvent
			}

			connector.eventChan <- botEvent
		case <-connector.errorChan:
			log.Print("disconnect server")
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
	_, err := connector.PostMessage(event.Channel, text, "")
	if err != nil {
		return err
	}

	return nil
}

func (connector *Connector) SendWithConfirm(event *bot.Event, username, text string) (string, error) {
	res, err := connector.PostMessage(event.Channel, text, username)
	if err != nil {
		return "", err
	}

	return res.Ts, nil
}

func (connector *Connector) SendPrivate(event *bot.Event, userId, text string) error {
	channel, err := connector.IMOpen(userId)
	if err != nil {
		return err
	}

	_, err = connector.PostMessage(channel.Id, text, "")
	return err
}

func (connector *Connector) Attach(event *bot.Event, fileName string, file io.Reader, title string) error {
	buf := bytes.NewBuffer([]byte{})
	w := multipart.NewWriter(buf)
	f, err := w.CreateFormField("token")
	if err != nil {
		return err
	}
	f.Write([]byte(connector.token))
	f, err = w.CreateFormField("channels")
	if err != nil {
		return err
	}
	f.Write([]byte(event.Channel))
	f, err = w.CreateFormField("title")
	if err != nil {
		return err
	}
	f.Write([]byte(title))
	f, err = w.CreateFormField("filetype")
	if err != nil {
		return err
	}
	f.Write([]byte("auto"))
	f, err = w.CreateFormField("filename")
	if err != nil {
		return err
	}
	f.Write([]byte(fileName))
	f, err = w.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	io.Copy(f, file)
	w.Close()

	req, err := http.NewRequest("POST", "https://slack.com/api/files.upload", buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.New("Failed file upload")
	}

	dec := json.NewDecoder(res.Body)
	var data struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}
	dec.Decode(&data)
	if data.Ok == false {
		return errors.New(data.Error)
	}

	return nil
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
				connector.sendTyping(c)
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
		info, err := connector.GetTeamInfo()
		if err != nil {
			return ""
		}
		connector.domain = info.Domain
	}

	return connector.domain
}

func (connector *Connector) startReading() {
	log.Print("start reading")
	var msg []byte
	connector.bufChan = make(chan []byte)

	go func() {
		tmp := make([]byte, ReadBufferSize)
		for {
			connector.connection.SetReadDeadline(time.Now().Add(ReadTimeout))
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
		log.Print("failed json.Marshal")
		return err
	}

	connector.connection.SetWriteDeadline(time.Now().Add(WriteTimeout))
	_, err = connector.connection.Write(buf)
	if err != nil {
		log.Print("failed send ping")
		return err
	}

	return nil
}

func (connector *Connector) sendTyping(channel string) error {
	typing := &Typing{Id: 1, Type: "typing", Channel: channel}
	buf, err := json.Marshal(typing)
	if err != nil {
		log.Print("faild json.Marshal")
		return err
	}

	connector.connection.SetWriteDeadline(time.Now().Add(WriteTimeout))
	_, err = connector.connection.Write(buf)
	if err != nil {
		return err
	}

	return nil
}
