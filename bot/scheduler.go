package bot

import (
	"context"
	"errors"
	"time"
)

const (
	interval = 1 * time.Minute
)

var (
	ErrIntervalLessThanMinute = errors.New("scheduler: interval must not be less than 1 minute")
)

type UnitTime int

const (
	Hourly UnitTime = iota
	Daily
)

type ScheduleFunc func(event *Event)

type ScheduleEntry struct {
	Channel string

	interval time.Duration
	next     time.Time
	f        ScheduleFunc
}

type Scheduler struct {
	entries   []*ScheduleEntry
	ctx       context.Context
	cancel    context.CancelFunc
	eventChan chan *ScheduleEntry
}

func NewScheduler() *Scheduler {
	return &Scheduler{entries: make([]*ScheduleEntry, 0), eventChan: make(chan *ScheduleEntry)}
}

func (scheduler *Scheduler) Start(ctx context.Context) {
	c, cancelFunc := context.WithCancel(ctx)
	scheduler.ctx = c
	scheduler.cancel = cancelFunc

	timer := time.NewTicker(interval)
	defer timer.Stop()

SchedulerLoop:
	for {
		select {
		case <-timer.C:
			t := time.Now()
			for _, entry := range scheduler.entries {
				if entry.CanExecute(t) {
					scheduler.eventChan <- entry
				}
			}
		case <-scheduler.ctx.Done():
			break SchedulerLoop
		}
	}

	scheduler.Stop()
}

func (scheduler *Scheduler) Stop() {
	if scheduler.cancel == nil {
		return
	}

	scheduler.cancel()
}

func (scheduler *Scheduler) Every(interval time.Duration, channel string, f ScheduleFunc) error {
	return scheduler.addEntry(interval, channel, f)
}

func (scheduler *Scheduler) At(every UnitTime, hour, minute int, channel string, f ScheduleFunc) error {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}

	entry := &ScheduleEntry{Channel: channel, f: f}
	switch every {
	case Hourly:
		next := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, 0, 0, loc)
		if now.After(next) {
			next = next.Add(time.Hour)
		}
		entry.interval = time.Hour
		entry.next = next
	case Daily:
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		if now.After(next) {
			next = next.Add(24 * time.Hour) // Add a day
		}
		entry.interval = 24 * time.Hour
		entry.next = next
	}
	scheduler.entries = append(scheduler.entries, entry)

	return nil
}

func (scheduler *Scheduler) TriggeredEvent() chan *ScheduleEntry {
	return scheduler.eventChan
}

func (scheduler *Scheduler) addEntry(interval time.Duration, channel string, f ScheduleFunc) error {
	if interval < time.Minute {
		return ErrIntervalLessThanMinute
	}

	scheduler.entries = append(scheduler.entries, &ScheduleEntry{interval: interval, next: time.Now().Add(interval), Channel: channel, f: f})
	return nil
}

func (entry *ScheduleEntry) CanExecute(t time.Time) bool {
	if entry.next.Before(t) || entry.next.Equal(t) {
		return true
	}
	return false
}

func (entry *ScheduleEntry) Execute(msg *Event) {
	entry.next = time.Now().Add(entry.interval)
	entry.f(msg)
}

func (entry *ScheduleEntry) ToEvent() *Event {
	return &Event{Type: ScheduledEvent, Channel: entry.Channel}
}
