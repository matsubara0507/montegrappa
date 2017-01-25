package bot

import (
	"sync"
)

type TestConnector struct {
	SendMessages []string
	sync         sync.RWMutex
}

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

func (c *TestConnector) GetChannelInfo(channel string) (*ChannelInfo, error) {
	ci := &ChannelInfo{
		Id:   channel,
		Name: channel,
	}
	return ci, nil
}

func (c *TestConnector) Send(_ *Event, _username string, text string) error {
	c.sync.Lock()
	defer c.sync.Unlock()
	c.SendMessages = append(c.SendMessages, text)
	return nil
}

func (c *TestConnector) SendWithConfirm(_ *Event, _username string, text string) (string, error) {
	c.sync.Lock()
	defer c.sync.Unlock()
	c.SendMessages = append(c.SendMessages, text)
	return "", nil
}

func (c *TestConnector) Async() bool {
	return true
}

func (c *TestConnector) Idle() chan bool {
	return make(chan bool)
}
