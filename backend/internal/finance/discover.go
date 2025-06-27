package finance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type discoverCompetitorRates struct {
	CompetitorRates []discoverCompetitorRate `json:"competitorRates"`
}
type discoverCompetitorRate struct {
	Institutionname  string  `json:"institutionname"`
	Instrumentname   string  `json:"instrumentname"`
	Minimum          int     `json:"minimum"`
	Maximum          float64 `json:"maximum"`
	Shortdescription string  `json:"shortdescription"`
	Apr              float64 `json:"apr"`
	Apy              float64 `json:"apy"`
	Featured         bool    `json:"featured"`
	Discover         bool    `json:"discover"`
}

type discoverLegacyRates struct {
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

type discover struct{}

func (discover *discover) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: discover.getName(), CDs: []CD{}}

	url := "https://www.discoverbank.com/rates/competitor/snapshot.json?=&=&aff=NAT&product=004"
	resp, err := client.Get(url)
	if err != nil {
		return account, err
	}

	if resp.StatusCode != http.StatusOK {
		return account, errors.New("unexpected status code: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return account, err
	}

	var discoverRates discoverCompetitorRates
	if err = json.Unmarshal(body, &discoverRates); err != nil {
		return account, err
	}

	for _, rate := range discoverRates.CompetitorRates {
		if rate.Institutionname != "Discover Bank" {
			continue
		}

		term, err := strconv.Atoi(rate.Shortdescription)
		if err != nil {
			continue
		}

		cd := CD{Term: term, Rate: strconv.FormatFloat(rate.Apy, 'f', 2, 64)}
		account.CDs = append(account.CDs, cd)
	}

	if len(account.CDs) <= 0 {
		return account, errors.New("no CDs found")
	}

	account.sortCdsByTerm()
	return account, nil
}

func (discover *discover) getName() string {
	return "Discover"
}

func (discover *discover) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	account := &SavingsAccount{Name: discover.getName()}

	url := "https://www.discoverbank.com/rates/legacy/featured.json?=&=&aff=NAT"
	resp, err := client.Get(url)
	if err != nil {
		return account, err
	}

	if resp.StatusCode != http.StatusOK {
		return account, errors.New("unexpected status code: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return account, err
	}

	var data discoverLegacyRates
	if err = json.Unmarshal(body, &data); err != nil {
		return account, err
	}

	for _, instrument := range data.Dataset {
		if instrument.InstrumentName != "ONLINE SAVINGS" {
			continue
		}

		account := &SavingsAccount{
			Name: discover.getName(),
			Rate: strconv.FormatFloat(instrument.Apy, 'f', 2, 64),
		}
		return account, nil
	}

	return account, fmt.Errorf("no savings rate found")
}
