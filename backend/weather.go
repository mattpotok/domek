package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type OpenWeatherCurrentWeather struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  int     `json:"sea_level"`
		GrndLevel int     `json:"grnd_level"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// TODO return the weather
func getCurrentWeather() *OpenWeatherCurrentWeather {
	openWeatherApiKey := "8d9e82c58e8b61001c721cd6b037c39e"
	latitude, longitude := "42.4154", "-71.1565"

	request := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s&units=metric",
		latitude, longitude, openWeatherApiKey,
	)

	response, err := http.Get(request)
	if err != nil {
		log.Fatalf("Error fetching weather - %s", err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Error fetching weather - %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading response body - %s", err)
	}

	var currentWeather OpenWeatherCurrentWeather
	if err := json.Unmarshal(body, &currentWeather); err != nil {
		log.Fatalf("Error unmarshalling response body - %s", err)
	}

	fmt.Printf("%+v\n", currentWeather)
	fmt.Printf("Temperature: %f\n", currentWeather.Main.Temp)

	return &currentWeather
}
