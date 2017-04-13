package main

import (
	"os"
	"time"

	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	BotName := "debug"
	Team := "debug"
	IgnoreUsers := make([]string, 0)
	AcceptUsers := make([]string, 0)

	connector := slack.NewSlackConnector(Team, Token)
	robot := bot.NewBot(connector, BotName, IgnoreUsers, AcceptUsers)
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
	robot.Start()
}
