package finance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

/* Request to get data from Capital One
curl 'https://api.capitalone.com/deposits/products/~/search' -X POST -H 'Accept: application/json;v=4' -H 'Content-Type: application/json' --data-raw '{"include":["RATES"],"isRenewableRate":true}'
*/

type CapitalOneProducts struct {
	Products []struct {
		ProductDefinition struct {
			Channel                 string `json:"channel"`
			EndDate                 any    `json:"endDate"`
			Platform                string `json:"platform"`
			BeginDate               string `json:"beginDate"`
			ApyTierType             string `json:"apyTierType"`
			ProductCode             string `json:"productCode"`
			ProductName             string `json:"productName"`
			ProductType             string `json:"productType"`
			ProductGroup            string `json:"productGroup"`
			ProductLineID           string `json:"productLineId"`
			LineOfBusiness          string `json:"lineOfBusiness"`
			PilotStartDate          string `json:"pilotStartDate"`
			ProductTypeCode         string `json:"productTypeCode"`
			ProductClassCode        string `json:"productClassCode"`
			ProductShortName        string `json:"productShortName"`
			BackbookStartDate       any    `json:"backbookStartDate"`
			ServicingPlatform       string `json:"servicingPlatform"`
			FrontbookStartDate      string `json:"frontbookStartDate"`
			ProductDescription      string `json:"productDescription"`
			ProductDescriptionLong  string `json:"productDescriptionLong"`
			ProductTypeDescription  string `json:"productTypeDescription"`
			ProductClassDescription string `json:"productClassDescription"`
			RatesPrecision          struct {
				Margin                    int `json:"margin"`
				Interest                  int `json:"interest"`
				DailyPercentageRate       int `json:"dailyPercentageRate"`
				AnnualizedPercentageRate  int `json:"annualizedPercentageRate"`
				AnnualizedPercentageYield int `json:"annualizedPercentageYield"`
			} `json:"ratesPrecision"`
			ProductStatus string `json:"productStatus"`
		} `json:"productDefinition"`
		Rates []struct {
			RateEffectiveDate time.Time `json:"rateEffectiveDate"`
			RateType          string    `json:"rateType"`
			Region            string    `json:"region"`
			RateStatus        string    `json:"rateStatus"`
			RateTiers         []struct {
				Term                      any     `json:"term"`
				Margin                    any     `json:"margin"`
				Interest                  float64 `json:"interest"`
				MaximumBalance            float64 `json:"maximumBalance"`
				MinimumBalance            float64 `json:"minimumBalance"`
				DailyPercentageRate       float64 `json:"dailyPercentageRate"`
				AnnualizedPercentageRate  float64 `json:"annualizedPercentageRate"`
				AnnualizedPercentageYield float64 `json:"annualizedPercentageYield"`
			} `json:"rateTiers"`
		} `json:"rates"`
	} `json:"products"`
	LastCachedAt time.Time `json:"lastCachedAt"`
}

type CapitalOne struct{}

// TODO
func (cof *CapitalOne) getCdRates() ([]CD, error) {
	return nil, nil
}

func (cof *CapitalOne) getName() string {
	return "Capital One"
}

func (cof *CapitalOne) getSavingsRate() (string, error) {
	client := &http.Client{}
	url := "https://api.capitalone.com/deposits/products/~/search"
	jsonData := `{"include": ["RATES"], "isRenewableRate": true}`

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json;v=4")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var cofProducts CapitalOneProducts
	if err = json.Unmarshal(body, &cofProducts); err != nil {
		return "", err
	}

	for _, product := range cofProducts.Products {
		if product.ProductDefinition.ProductName != "360 Performance Savings" {
			continue
		}

		if len(product.Rates) <= 0 {
			return "", fmt.Errorf("no rates found")
		}

		if len(product.Rates[0].RateTiers) <= 0 {
			return "", fmt.Errorf("no rate tiers found")
		}

		return strconv.FormatFloat(product.Rates[0].RateTiers[0].AnnualizedPercentageYield, 'f', 2, 64), nil
	}

	return "", fmt.Errorf("unable to get savings rate")
}
