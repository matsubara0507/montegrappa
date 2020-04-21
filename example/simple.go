package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/f110/montegrappa/bot"
	slackConnector "github.com/f110/montegrappa/slack"
	"github.com/nlopes/slack"
)

func main() {
	Token := os.Getenv("SLACK_TOKEN")
	BotName := "debug"
	Team := "debug"
	ScheduleChannel := "C056M677R"
	IgnoreUsers := make([]string, 0)
	AcceptUsers := make([]string, 0)

	connector := slackConnector.NewConnector(Team, Token)
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
		connector, ok := msg.Bot.Connector.(*slackConnector.Connector)
		if !ok {
			return
		}

		cursor := ""
		joinedChannel := make([]slack.Channel, 0)
		for {
			channels, nextCursor, err := connector.Client().GetConversations(&slack.GetConversationsParameters{Cursor: cursor})
			if err != nil {
				log.Print(err)
				return
			}
			if len(channels) == 0 {
				break
			}

			for _, v := range channels {
				if v.IsMember {
					joinedChannel = append(joinedChannel, v)
				}
			}

			if nextCursor == "" {
				break
			} else {
				cursor = nextCursor
			}
		}

		for _, c := range joinedChannel {
			log.Printf("%s - %s", c.Name, c.ID)
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
