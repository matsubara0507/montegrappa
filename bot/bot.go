package bot

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"
)

const (
	defaultConnectBackoffTime = 5 // seconds
	maxConnectRetryCount      = 10
)

var (
	ErrFailedConnect = errors.New("failed connect")
)

type OnError func(*Event)

type Bot struct {
	Connector   Connector
	Name        string
	Persistence Persistence

	connectErrorChan  chan error
	eventHandler      *EventHandler
	scheduler         *Scheduler
	connectRetryCount int
	disconnectCount   int
	ctx               context.Context
	cancel            context.CancelFunc
}

func NewBot(connector Connector, persistence Persistence, name string, ignoreUsers []string, acceptUsers []string) *Bot {
	if persistence == nil {
		persistence = &NoneDB{}
	}
	return &Bot{
		Connector:         connector,
		Name:              name,
		Persistence:       persistence,
		connectErrorChan:  make(chan error),
		eventHandler:      NewEventHandler(ignoreUsers, acceptUsers),
		scheduler:         NewScheduler(),
		connectRetryCount: 0,
		disconnectCount:   0,
	}
}

func (bot *Bot) Start(ctx context.Context) error {
	c, cancel := context.WithCancel(ctx)
	bot.ctx = c
	bot.cancel = cancel

	go bot.scheduler.Start(bot.ctx)

	for {
		for {
			err := bot.Connect()
			if err != nil {
				bot.connectRetryCount += 1
				sleepTime := time.Duration(bot.connectRetryCount*defaultConnectBackoffTime) * time.Second
				time.Sleep(sleepTime)
				continue
			} else {
				bot.connectRetryCount = 0
				break
			}

			if bot.connectRetryCount > maxConnectRetryCount {
				return ErrFailedConnect
			}
		}

	RECEIVE:
		for {
			select {
			case event := <-bot.Connector.ReceivedEvent():
				event.Bot = bot
				if bot.Connector.Async() == true {
					go bot.eventHandler.Handle(event, true)
				} else {
					bot.eventHandler.Handle(event, false)
					bot.Connector.Idle() <- true
				}
			case entry := <-bot.scheduler.TriggeredEvent():
				e := entry.ToEvent()
				e.Bot = bot
				if bot.Connector.Async() {
					go entry.Execute(e)
				} else {
					entry.Execute(e)
					bot.Connector.Idle() <- true
				}
			case err := <-bot.connectErrorChan:
				bot.disconnectCount++
				log.Printf("reconnect: %s", err)
				break RECEIVE
			case <-ctx.Done():
				bot.Shutdown()
				return nil
			}
		}
	}

	return nil
}

func (bot *Bot) Shutdown() error {
	bot.Persistence.Close()
	bot.cancel()
	return nil
}

func (bot *Bot) Connect() error {
	err := bot.Connector.Connect()
	if err != nil {
		return err
	}

	go func() {
		err := bot.Connector.Listen()
		if err != nil {
			bot.connectErrorChan <- err
		}
	}()

	return nil
}

func (bot *Bot) OnError(f OnError) {
	bot.eventHandler.OnError = f
}

func (bot *Bot) Send(event *Event, text string) {
	bot.Connector.Send(event, bot.Name, text)
}

func (bot *Bot) Sendf(event *Event, format string, a ...interface{}) {
	text := fmt.Sprintf(format, a...)
	bot.Send(event, text)
}

func (bot *Bot) SendWithConfirm(event *Event, text, reaction string, callback func(*Event)) {
	id, _ := bot.Connector.SendWithConfirm(event, bot.Name, text)
	bot.eventHandler.RequireReaction(event.Channel, id, reaction, event.User.Id, callback)
}

func (bot *Bot) SendWithConfirmf(event *Event, reaction string, callback func(*Event), format string, a ...interface{}) {
	text := fmt.Sprintf(format, a)
	bot.SendWithConfirm(event, text, reaction, callback)
}

func (bot *Bot) SendRequireResponse(event *Event, text string) (func(), chan string) {
	bot.Connector.Send(event, bot.Name, text)
	return bot.eventHandler.RequireResponse(event.Channel, event.User.Id)
}

func (bot *Bot) SendRequireResponsef(event *Event, format string, a ...interface{}) (func(), chan string) {
	text := fmt.Sprintf(format, a)
	return bot.SendRequireResponse(event, text)
}

func (bot *Bot) WithIndicate(channel string, f func() error) {
	cancel := bot.Connector.WithIndicate(channel)
	defer cancel()
	f()
}

func (bot *Bot) Attach(event *Event, title, fileName string, file io.Reader) error {
	return bot.Connector.Attach(event, fileName, file, title)
}

func (bot *Bot) SendPrivate(event *Event, text string) {
	bot.Connector.SendPrivate(event, event.User.Id, text)
}

func (bot *Bot) GetPermalink(event *Event) string {
	return bot.Connector.GetPermalink(event)
}

func (bot *Bot) Hear(pattern string, callback func(*Event)) {
	bot.eventHandler.AddCommand(regexp.MustCompile(pattern), "", callback, false)
}

func (bot *Bot) Command(pattern string, description string, callback func(*Event)) {
	bot.eventHandler.AddCommand(regexp.MustCompile("\\A"+bot.Name+"\\s+"+pattern+"\\z"), pattern+" - "+description, callback, false)
}

func (bot *Bot) CommandWithArgv(pattern string, description string, callback func(*Event)) {
	bot.eventHandler.AddCommand(regexp.MustCompile("\\A"+bot.Name+"\\s+"+pattern+"(?:\\s+(.+))*\\z"), pattern+" - "+description, callback, true)
}

func (bot *Bot) Appearance(user string, callback func(*Event)) {
	bot.eventHandler.Appearance(user, callback)
}

func (bot *Bot) Every(interval time.Duration, channel string, callback ScheduleFunc) {
	if err := bot.scheduler.Every(interval, channel, callback); err != nil {
		panic(err)
	}
}

func (bot *Bot) At(every Every, hour, minute int, channel string, f ScheduleFunc) {
	if err := bot.scheduler.At(every, hour, minute, channel, f); err != nil {
		panic(err)
	}
}
