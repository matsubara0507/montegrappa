package bot

var (
	initializer = make([]HandlerInitializer, 0)
)

type HandlerInitializer func(bot *Bot)

func DefineHandler(i HandlerInitializer) {
	initializer = append(initializer, i)
}

func Init(bot *Bot) {
	for _, f := range initializer {
		f(bot)
	}
}
