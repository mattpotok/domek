package main

// TODO move this file to /cmd/domek

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mattpotok/domek/backend/internal/common"
	"github.com/mattpotok/domek/backend/internal/database"
	"github.com/mattpotok/domek/backend/internal/telegram"
	"github.com/playwright-community/playwright-go"
)

// TODO migrate to its own file
func initialize_database() *sql.DB {
	db, err := sql.Open("sqlite3", "domek.db")
	if err != nil {
		log.Fatalf("Error opening database - %s", err)
	}

	weatherTable := `
		CREATE TABLE IF NOT EXISTS weather (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			datetime    TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			latitude    REAL NOT NULL,
			longitude   REAL NOT NULL,
			humidity    REAL NOT NULL,
			pressure    REAL NOT NULL,
			temperature REAL NOT NULL
		);`
	_, err = db.Exec(weatherTable)
	if err != nil {
		log.Fatalf("Error creating weather table - %s", err)
	}

	return db
}

func insertWeather(db *sql.DB, weather *OpenWeatherCurrentWeather) {
	query := "INSERT INTO weather (latitude, longitude, humidity, pressure, temperature) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, weather.Coord.Lat, weather.Coord.Lon, weather.Main.Humidity, weather.Main.Pressure, weather.Main.Temp)
	if err != nil {
		log.Fatalf("Error inserting weather data - %s", err)
	}
}

func getWeather(db *sql.DB, start time.Time, end time.Time) {
	query := "SELECT datetime, temperature FROM weather WHERE datetime BETWEEN ? AND ?"
	rows, err := db.Query(query, start, end)
	if err != nil {
		log.Fatalf("Error querying weather data - %s", err)
	}

	for rows.Next() {
		var datetime string
		var temperature float64
		err := rows.Scan(&datetime, &temperature)
		if err != nil {
			log.Fatalf("Error scanning weather data - %s", err)
		}

		fmt.Println(datetime, temperature)
	}
}

func main() {
	// TODO pass this around to other places
	common.CreateServiceDirectory()
	cfg := common.LoadConfig()

	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Error initializing database - %s", err)
	}

	// weather := getCurrentWeather()
	// insertWeather(db, weather)

	// start := time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC)
	// end := time.Date(2024, 12, 30, 1, 0, 0, 0, time.UTC)

	// getWeather(db, start, end)

	// db.Close()

	// TODO

	/*
		var rate float64
		_, err := fmt.Sscanf("/finance_effective_rates 1", "/finance_effective_rates %f", &rate)
		if err != nil {
			log.Fatalf("Error parsing rate - %s", err)
		}

		fmt.Println(rate)
	*/

	ctx := context.Background()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	initialize_playwright()

	_, err = telegram.NewTelegram(ctx, cfg.Telegram, db)
	if err != nil {
		log.Fatalf("Unable to initialize Telegram - %s", err)
	}

	start()
}

func initialize_playwright() {
	// TODO check if one needs to install playwright here or not and log a message
	log.Println("Installing Playwright...")
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}, OnlyInstallShell: true})
	if err != nil {
		log.Fatalf("Error installing playwright - %s", err)
	}
	log.Println("Playwright has been installed/updated.")
}

func start() {
	/*
		region := getEnvironmentVariable(REGION_ENV)
		snsTopicArn := getEnvironmentVariable(SNS_TOPIC_ARN_ENV)

		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("domek"))
		if err != nil {
			log.Fatalf("Unable to load AWS configuration - %s", err)
		}

		snsActions := NewSnsActions(cfg, region)
		notifier := NewSnsEmailNotifier(snsActions, snsTopicArn)
	*/

	// scheduler := cron.New()
	// scheduler.AddFunc("0 "

	/*
		ticker := NewTicker("19:00:00", notifier)
		go ticker.Run()

		// TODO move this out into `api.go` and shove into a go thread
		ctrl := &Controller{
			Notifier: notifier,
		}

		http.HandleFunc("POST /events", ctrl.PostEvent)
		http.HandleFunc("GET /finance/cds", ctrl.GetFinanceCDs)
	*/

	// TODO look up how to do this in a go thread
	err := http.ListenAndServe(":3333", nil)
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

// FIXME
type CDRates struct {
	Institution       string `json:"institution"`
	ThreeMonthRate    string `json:"three_month_rate"`
	SixMonthRate      string `json:"six_month_rate"`
	NineMonthRate     string `json:"nine_month_rate"`
	TwelveMonthRate   string `json:"twelve_month_rate"`
	EighteenMonthRate string `json:"eighteen_month_rate"`
}

// FIXME
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
