package handler

import (
	"github.com/f110/montegrappa/bot"
)

type Cmd struct {
	Pattern     string
	Description string
	Handler     func(*bot.Event)
}

var commands = make([]*Cmd, 0)

func Init(bot *bot.Bot) {
	for _, c := range commands {
		bot.Command(c.Pattern, c.Description, c.Handler)
	}
}

func AddCommand(pattern, description string, handler func(*bot.Event)) {
	commands = append(commands, &Cmd{Pattern: pattern, Description: description, Handler: handler})
}

func ShowHelp() string {
	help := ""
	for _, c := range commands {
		if c.Description != "" {
			help += c.Pattern + ": " + c.Description + "\n"
		}
	}

	return help
}
