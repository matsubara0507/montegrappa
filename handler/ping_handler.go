package handler

import (
	"github.com/f110/montegrappa/bot"
)

func init() {
	bot.DefineHandler(func(bot *bot.Bot) {
		bot.Command("ping", "ピンポン", PingHandler)
	})
}

func PingHandler(msg *bot.Event) {
	msg.Say("pong")
}
