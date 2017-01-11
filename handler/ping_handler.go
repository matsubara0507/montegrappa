package handler

import (
	"github.com/f110/montegrappa/bot"
)

func init() {
	AddCommand("ping", "ピンポン", PingHandler)
}

func PingHandler(msg *bot.Event) {
	msg.SayWithConfirm("pongって言って欲しいなら :ok_hand: をつけてください", "ok_hand", func(msg *bot.Event) {
		msg.Say("pong")
	})
}
