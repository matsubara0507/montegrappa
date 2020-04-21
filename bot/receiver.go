package bot

type Receiver interface {
	ReceivedEvent() chan *Event
	Setup() error
	Start() error
}
