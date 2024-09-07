package main

import (
	"log"
	"time"
)

type Ticker struct {
	sched    time.Time
	notifier Notifier
	timer    *time.Timer
}

func NewTicker(when string, notifier Notifier) *Ticker {
	sched, err := time.Parse(time.TimeOnly, when)
	if err != nil {
		panic("Unable to parse 'when'")
	}

	return &Ticker{
		sched:    sched,
		notifier: notifier,
		timer:    nil,
	}
}

func (t *Ticker) Run() {
	if t.timer != nil {
		t.timer.Stop()
	}

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), t.sched.Hour(), t.sched.Minute(), t.sched.Second(), 0, time.Local)
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		diff := time.Until(next)

		log.Printf("Next tick at '%+v'\n", next)
		if t.timer == nil {
			t.timer = time.NewTimer(diff)
		} else {
			t.timer.Reset(diff)
		}

		<-t.timer.C

		event := Event{
			What:  "timer",
			Where: "domek",
			Why:   "tick",
		}
		t.notifier.Notify(event)
	}
}
