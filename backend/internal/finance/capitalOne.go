package finance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mattpotok/domek/backend/internal/common"
)

type capitalOneProducts struct {
	Products     []capitalOneProduct `json:"products"`
	LastCachedAt time.Time           `json:"lastCachedAt"`
}

type capitalOneProduct struct {
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
}

type capitalOne struct{}

func getCapitalOneProducts(client *http.Client) (*capitalOneProducts, error) {
	url := "https://api.capitalone.com/deposits/products/~/search"
	jsonData := `{"include": ["RATES"], "isRenewableRate": true}`

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json;v=4")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var products capitalOneProducts
	if err = json.Unmarshal(body, &products); err != nil {
		return nil, err
	}

	return &products, nil
}

func (capitalOne *capitalOne) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: capitalOne.getName(), CDs: []CD{}}

	capitalOneProducts, err := getCapitalOneProducts(client)
	if err != nil {
		return account, err
	}

	product, found := common.Find(
		capitalOneProducts.Products,
		func(product capitalOneProduct) bool { return product.ProductDefinition.ProductName == "CD" },
	)
	if !found {
		return account, errors.New("product 'CD' was not found")
	}

	if len(product.Rates) != 1 {
		return account, errors.New("expected 'rates' to be of length 1")
	}

	if len(product.Rates[0].RateTiers) <= 0 {
		return account, errors.New("no 'rate tiers' found")
	}

	for _, rateTier := range product.Rates[0].RateTiers {
		term, ok := rateTier.Term.(float64)
		if !ok {
			continue
		}

		cd := CD{
			Term: int(term),
			Rate: strconv.FormatFloat(rateTier.AnnualizedPercentageYield, 'f', 2, 64),
		}
		account.CDs = append(account.CDs, cd)
	}

	if len(account.CDs) <= 0 {
		return account, errors.New("no CDs found")
	}

	account.sortCdsByTerm()
	return account, nil
}

func (capitalOne *capitalOne) getName() string {
	return "Capital One"
}

func (capitalOne *capitalOne) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	capitalOneProducts, err := getCapitalOneProducts(client)
	if err != nil {
		return nil, err
	}

	idx := -1
	for i, product := range capitalOneProducts.Products {
		if product.ProductDefinition.ProductName != "360 Performance Savings" {
			idx = i
			break
		}
	}

	if idx < 0 {
		return nil, errors.New("product '360 Performance Savings' was not found")
	}

	for _, product := range capitalOneProducts.Products {
		if product.ProductDefinition.ProductName != "360 Performance Savings" {
			continue
		}

		if len(product.Rates) <= 0 {
			return nil, fmt.Errorf("no rates found")
		}

		if len(product.Rates[0].RateTiers) <= 0 {
			return nil, fmt.Errorf("no rate tiers found")
		}

		account := &SavingsAccount{
			Name: capitalOne.getName(),
			Rate: strconv.FormatFloat(product.Rates[0].RateTiers[0].AnnualizedPercentageYield, 'f', 2, 64),
		}
		return account, nil
	}

	return nil, fmt.Errorf("unable to get savings rate")
}
