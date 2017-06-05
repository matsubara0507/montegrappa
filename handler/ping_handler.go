package handler

import (
	"github.com/f110/montegrappa/bot"
)

func init() {
	bot.AddCommand("ping", "ピンポン", PingHandler)
}

func PingHandler(msg *bot.Event) {
	msg.Say("pong")
}
