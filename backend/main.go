package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	region := getEnvironmentVariable(REGION_ENV)
	snsTopicArn := getEnvironmentVariable(SNS_TOPIC_ARN_ENV)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("domek"))
	if err != nil {
		log.Fatalf("Unable to load AWS configuration - %s", err)
	}

	snsActions := NewSnsActions(cfg, region)
	notifier := NewSnsEmailNotifier(snsActions, snsTopicArn)

	ticker := NewTicker("19:00:00", notifier)
	go ticker.Run()

	ctrl := &Controller{
		Notifier: notifier,
	}

	http.HandleFunc("/events", ctrl.PostEvent)

	err = http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatalf("Error starting server - %s", err)
	}
}

// A Controller injects dependencies into request handling
type Controller struct {
	Notifier Notifier
}

func (ctrl *Controller) PostEvent(w http.ResponseWriter, r *http.Request) {
	var event Event
	err := decodeJSONBody(w, r, &event)
	if err != nil {
		var mr *malformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	log.Printf("POST /event - %v\n", event)

	ctrl.Notifier.Notify(event)
}
