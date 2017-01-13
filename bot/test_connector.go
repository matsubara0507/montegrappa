package bot

type TestConnector struct {}

func NewTestConnector() *TestConnector {
	return &TestConnector{}
}

func (c *TestConnector) Connect() {
}

func (c *TestConnector) Listen() error {
	return nil
}

func (c *TestConnector) ReceivedEvent() chan *Event {
	return make(chan *Event)
}

func (c *TestConnector) Send(_ *Event, _ string, _ string) error {
	return nil
}

func (c *TestConnector) SendWithConfirm(_ *Event, _ string, _ string) (string, error) {
	return "", nil
}

func (c *TestConnector) Async() bool {
	return true
}

func (c *TestConnector) Idle() chan bool {
	return make(chan bool)
}
