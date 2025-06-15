package finance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type AllyRates struct {
	CompetitorProdRates struct {
		Rates []struct {
			RateID      string `json:"rateId"`
			ProductCode string `json:"productCode"`
			Banks       []struct {
				BankName      string `json:"bankName"`
				EffectiveDate string `json:"effectiveDate"`
				Terms         []struct {
					TermID   string `json:"termId"`
					Term     int    `json:"term"`
					TermCode string `json:"termCode"`
					Rates    []struct {
						Max float64 `json:"max"`
						Apr float64 `json:"apr"`
						Apy float64 `json:"apy"`
						Min float64 `json:"min"`
					} `json:"rates"`
					TermStartDate any `json:"termStartDate"`
					TermEndDate   any `json:"termEndDate"`
				} `json:"terms"`
			} `json:"banks"`
		} `json:"rates"`
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"competitorProdRates"`
}

type Ally struct{}

func (ally *Ally) getCdRates() ([]CD, error) {
	return nil, nil
}

func (ally *Ally) getName() string {
	return "Ally"
}

func (ally *Ally) getSavingsRate() (string, error) {
	client := &http.Client{}
	url := "https://secure.ally.com/acs//products/temporary/competitor?format=json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("api-key", "TMqqqyuyVeNnD1HTgvsNu6lulUHL9ksc")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data AllyRates
	if err = json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	for _, rates := range data.CompetitorProdRates.Rates {
		if rates.ProductCode != "OSAV" {
			continue
		}

		for _, bank := range rates.Banks {
			if bank.BankName != "Ally Bank" {
				continue
			}

			if len(bank.Terms) <= 0 {
				return "", fmt.Errorf("no terms found")
			}

			if len(bank.Terms[0].Rates) <= 0 {
				return "", fmt.Errorf("no rates found")
			}

			return strconv.FormatFloat(bank.Terms[0].Rates[0].Apy, 'f', 2, 64), nil
		}
	}

	return "", fmt.Errorf("unable to get savings rate")
}
