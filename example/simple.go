package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/slack"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	BotName := "debug"
	Team := "debug"
	ScheduleChannel := "C056M677R"
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
	robot.Command("channels", "channels", func(msg *bot.Event) {
		channels, err := msg.Bot.Connector.(*slack.Connector).GetJoinedChannelList()
		if err != nil {
			log.Print(err)
			return
		}
		for _, c := range channels {
			log.Printf("%s - %s", c.Name, c.Id)
		}
	})
	robot.WatchReaction("thumbsup", func(msg *bot.Event) {
		log.Print(msg)
	})
	robot.Every(1*time.Hour, ScheduleChannel, func(event *bot.Event) {
		event.Say("<!here> Hi")
	})
	robot.At(bot.Daily, 19, 46, "C056M677R", func(event *bot.Event) {
		event.Say("daily")
	})
	robot.Start(context.Background())
}
