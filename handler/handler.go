package handler

import (
	"github.com/f110/montegrappa/bot"
)

type Cmd struct {
	Type        string
	Pattern     string
	Description string
	Handler     func(*bot.Event)
}

const (
	typeCmdWithArgv = "cmd_with_argv"
	typeHear        = "hear"
)

var commands = make([]*Cmd, 0)

func Init(bot *bot.Bot) {
	for _, c := range commands {
		switch c.Type {
		case typeCmdWithArgv:
			bot.CommandWithArgv(c.Pattern, c.Description, c.Handler)
		case typeHear:
			bot.Hear(c.Pattern, c.Handler)
		default:
			bot.Command(c.Pattern, c.Description, c.Handler)
		}
	}
}

func AddCommand(pattern, description string, handler func(*bot.Event)) {
	commands = append(commands, &Cmd{Pattern: pattern, Description: description, Handler: handler})
}

func AddCommandWithArgv(pattern, description string, handler func(*bot.Event)) {
	commands = append(commands, &Cmd{Type: typeCmdWithArgv, Pattern: pattern, Description: description, Handler: handler})
}

func Hear(pattern string, handler func(*bot.Event)) {
	commands = append(commands, &Cmd{Type: typeHear, Pattern: pattern})
}

func ShowHelp() string {
	help := ""
	for _, c := range commands {
		if c.Type == typeHear {
			continue
		}
		if c.Description != "" {
			help += c.Pattern + ": " + c.Description + "\n"
		}
	}

	return help
}
