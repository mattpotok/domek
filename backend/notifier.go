package main

import (
	"fmt"
	"log"
	"time"
)

type Notifier interface {
	Notify(event Event)
}

const sns_max_daily_emails = 30

type SnsEmailNotifier struct {
	eventQueue     chan Event
	lastNotifyDate time.Time
	numEmailsSent  int
	pendingEvents  []Event
	snsActions     *SnsActions
	snsTopicArn    string
}

func NewSnsEmailNotifier(snsActions *SnsActions, snsTopicArn string) *SnsEmailNotifier {
	notifier := &SnsEmailNotifier{
		eventQueue:     make(chan Event, 16),
		lastNotifyDate: time.Now().AddDate(0, 0, -1),
		numEmailsSent:  0,
		pendingEvents:  []Event{},
		snsActions:     snsActions,
		snsTopicArn:    snsTopicArn,
	}

	go notifier.processEventQueue()

	return notifier
}

func (notifier *SnsEmailNotifier) Notify(event Event) {
	now := time.Now()
	event.When = now.Format(time.RFC1123)
	notifier.eventQueue <- event
}

func (notifier *SnsEmailNotifier) processEventQueue() {
	for {
		event := <-notifier.eventQueue
		notifier.processEvent(event)
	}
}

func (notifier *SnsEmailNotifier) processEvent(event Event) {
	log.Printf("Processing event - %+v\n", event)

	now := time.Now()
	if !areSameDate(now, notifier.lastNotifyDate) {
		notifier.lastNotifyDate = now
		notifier.numEmailsSent = 0
	}

	if event.What == "timer" && event.Where == "domek" && event.Why == "tick" {
		message := "List of pending events"
		for _, pendingEvent := range notifier.pendingEvents {
			message += fmt.Sprintf("\n- %s", pendingEvent.toNotifyString())
		}
		notifier.snsActions.Publish(notifier.snsTopicArn, message)
		notifier.pendingEvents = notifier.pendingEvents[:0]

		return
	}

	if notifier.numEmailsSent < sns_max_daily_emails {
		message := event.toNotifyString()
		_, err := notifier.snsActions.Publish(notifier.snsTopicArn, message)
		if err != nil {
			log.Printf("Unable to publish message - %v", err)
		} else {
			notifier.numEmailsSent += 1
		}
	} else {
		notifier.pendingEvents = append(notifier.pendingEvents, event)
	}
}

func areSameDate(t1 time.Time, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
