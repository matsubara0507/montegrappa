package bot

import "time"

type Cmd struct {
	Type        string
	Pattern     string
	Description string
	Handler     func(*Event)
}

const (
	typeCmdWithArgv = "cmd_with_argv"
	typeHear        = "hear"
)

type everyEntry struct {
	interval time.Duration
	channel  string
	callback ScheduleFunc
}

type atEntry struct {
	unitTime UnitTime
	hour     int
	minute   int
	channel  string
	callback ScheduleFunc
}

var (
	commands = make([]*Cmd, 0)
	every    = make([]*everyEntry, 0)
	at       = make([]*atEntry, 0)
)

func Init(bot *Bot) {
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

	for _, e := range every {
		bot.Every(e.interval, e.channel, e.callback)
	}

	for _, a := range at {
		bot.At(a.unitTime, a.hour, a.minute, a.channel, a.callback)
	}
}

func AddCommand(pattern, description string, handler func(*Event)) {
	commands = append(commands, &Cmd{Pattern: pattern, Description: description, Handler: handler})
}

func AddCommandWithArgv(pattern, description string, handler func(*Event)) {
	commands = append(commands, &Cmd{Type: typeCmdWithArgv, Pattern: pattern, Description: description, Handler: handler})
}

func Hear(pattern string, handler func(*Event)) {
	commands = append(commands, &Cmd{Type: typeHear, Pattern: pattern, Handler: handler})
}

func Every(interval time.Duration, channel string, callback ScheduleFunc) {
	every = append(every, &everyEntry{interval, channel, callback})
}

func At(every UnitTime, hour, minute int, channel string, callback ScheduleFunc) {
	at = append(at, &atEntry{every, hour, minute, channel, callback})
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
