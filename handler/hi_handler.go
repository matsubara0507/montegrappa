package handler

import (
	"github.com/f110/montegrappa/bot"
	"github.com/f110/montegrappa/db"
	"log"
	"strconv"
)

func init() {
	AddCommand("hi", "hi", ShowInstanceInfo)
}

func ShowInstanceInfo(msg *bot.Event) {
	info, err := db.ReadInstanceInfo()
	if err != nil {
		log.Print(err)
		return
	}
	msg.Say("私は " + strconv.Itoa(int(info.Seq)) + "人目ですね。" + info.StartAt.Format("2006年1月2日 15時04分") + "に生まれました。")
}
