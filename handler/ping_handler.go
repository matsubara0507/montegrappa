package handler

import (
	"github.com/matsubara0507/montegrappa/bot"
)

func init() {
	bot.DefineHandler(func(bot *bot.Bot) {
		bot.Command("ping", "ピンポン", PingHandler)
	})
}

func PingHandler(msg *bot.Event) {
	msg.Say("pong")
}
