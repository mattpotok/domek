package finance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type DiscoverSavingRates struct {
	Dataset []struct {
		Type           string  `json:"__type"`
		InstrumentName string  `json:"instrumentName"`
		Apy            float64 `json:"apy"`
		Disclaimer     string  `json:"disclaimer"`
		Disclaimer2    string  `json:"disclaimer2"`
		TierMinimum    int     `json:"tierMinimum"`
		TermName       string  `json:"termName"`
		TermLength     string  `json:"termLength"`
		ActiveAsOf     string  `json:"activeAsOf"`
	} `json:"dataset"`
}

type Discover struct{}

// TODO
func (discover *Discover) getCdRates() ([]CD, error) {
	return nil, nil
}

func (discover *Discover) getName() string {
	return "Discover"
}

// TODO do I need to pass in playwright or an HTTP client?
func (discover *Discover) getSavingsRate() (string, error) {
	url := "https://www.discoverbank.com/rates/legacy/featured.json?=&=&aff=NAT"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data DiscoverSavingRates
	if err = json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	for _, instrument := range data.Dataset {
		if instrument.InstrumentName != "ONLINE SAVINGS" {
			continue
		}

		return strconv.FormatFloat(instrument.Apy, 'f', 2, 64), nil
	}

	return "", fmt.Errorf("unable to get savings rate")
}
