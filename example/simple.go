package main

import (
	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
	"os"
	"time"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	ProfileIcon := ""
	BotName := "debug"
	IgnoreUsers := make([]string, 0)

	connector := slack.NewSlackConnector(Token, ProfileIcon)
	robot := bot.NewBot(connector, BotName, IgnoreUsers)
	robot.Command("ping", "ping pong", func(msg *bot.Event) {
		msg.Say("pong")
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
	robot.Start()
}
