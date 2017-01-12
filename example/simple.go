package main

import (
	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
	"os"
	"time"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	BotName := "debug"
	IgnoreUsers := make([]string, 0)

	connector := slack.NewSlackConnector(Token)
	robot := bot.NewBot(connector, BotName, IgnoreUsers)
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
