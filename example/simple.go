package main

import (
	"context"
	"os"
	"time"

	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	BotName := "debug"
	Team := "debug"
	ScheduleChannel := "debug"
	IgnoreUsers := make([]string, 0)
	AcceptUsers := make([]string, 0)

	connector := slack.NewConnector(Team, Token)
	robot := bot.NewBot(connector, nil, BotName, IgnoreUsers, AcceptUsers)
	robot.Command("ping", "ping pong", func(msg *bot.Event) {
		msg.Sayf("pong %l", msg.User)
	})
	robot.Command("test", "test", func(msg *bot.Event) {
		cancelFunc, resChan := msg.SayRequireResponse("どうしましたか？")
		go func() {
			t := time.Tick(1 * time.Minute)
			select {
			case <-t:
				cancelFunc()
			}
		}()
		res := <-resChan
		msg.Say(res)
	})
	robot.CommandWithArgv("test2", "test2", func(msg *bot.Event) {
		for _, v := range msg.Argv {
			msg.Say(v)
		}
	})
	robot.Every(1*time.Minute, ScheduleChannel, func(event *bot.Event) {
		event.Say("Hi")
	})
	robot.At(bot.Daily, 19, 46, "C056M677R", func(event *bot.Event) {
		event.Say("daily")
	})
	robot.Start(context.Background())
}
