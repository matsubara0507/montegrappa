package main

import (
	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
	"os"
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
	robot.Start()
}
