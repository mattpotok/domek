package main

// TODO move this file to /cmd/domek

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx := context.Background()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	region := getEnvironmentVariable(REGION_ENV)
	snsTopicArn := getEnvironmentVariable(SNS_TOPIC_ARN_ENV)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("domek"))
	if err != nil {
		log.Fatalf("Unable to load AWS configuration - %s", err)
	}

	snsActions := NewSnsActions(cfg, region)
	notifier := NewSnsEmailNotifier(snsActions, snsTopicArn)

	_, err = NewTelegramBot(ctx)
	if err != nil {
		log.Fatalf("Unable to initialize Telegram bot - %s", err)
	}

	// scheduler := cron.New()
	// scheduler.AddFunc("0 "

	ticker := NewTicker("19:00:00", notifier)
	go ticker.Run()

	// TODO move this out into `api.go` and shove into a go thread
	ctrl := &Controller{
		Notifier: notifier,
	}

	http.HandleFunc("POST /events", ctrl.PostEvent)
	http.HandleFunc("GET /finance/cds", ctrl.GetFinanceCDs)

	// TODO look up how to do this in a go thread
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
	log.Println("POST /event")

	var event Event
	err := decodeJSONBody(w, r, &event)
	if err != nil {
		log.Printf("Error processing event - %v\n", err)

		var mr *malformedRequest

		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	log.Printf("Received event - %v\n", event)

	ctrl.Notifier.Notify(event)

	// TODO return a success here
}

type CDRates struct {
	Institution       string `json:"institution"`
	ThreeMonthRate    string `json:"three_month_rate"`
	SixMonthRate      string `json:"six_month_rate"`
	NineMonthRate     string `json:"nine_month_rate"`
	TwelveMonthRate   string `json:"twelve_month_rate"`
	EighteenMonthRate string `json:"eighteen_month_rate"`
}

func (ctrl *Controller) GetFinanceCDs(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /finance/cds")

	cdRates := CDRates{
		Institution:       "Ally",
		ThreeMonthRate:    "3.00",
		SixMonthRate:      "4.40",
		NineMonthRate:     "4.30",
		TwelveMonthRate:   "4.25",
		EighteenMonthRate: "4.00",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cdRates)
}
