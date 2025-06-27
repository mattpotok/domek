package finance

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/mattpotok/domek/backend/internal/common"
)

type allyProducts struct {
	CompetitorProdRates struct {
		Rates   []allyCompetitorProduct `json:"rates"`
		Status  string                  `json:"status"`
		Message string                  `json:"message"`
	} `json:"competitorProdRates"`
}

type allyCompetitorProduct struct {
	RateID      string               `json:"rateId"`
	ProductCode string               `json:"productCode"`
	Banks       []allyCompetitorBank `json:"banks"`
}

type allyCompetitorBank struct {
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
}

// TODO this should be lowercase
type ally struct{}

func getAllyProducts(client *http.Client) (*allyProducts, error) {
	url := "https://secure.ally.com/acs//products/temporary/competitor?format=json"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("api-key", "TMqqqyuyVeNnD1HTgvsNu6lulUHL9ksc")
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

	var products allyProducts
	if err = json.Unmarshal(body, &products); err != nil {
		return nil, err
	}

	return &products, nil
}

func (ally *ally) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: ally.getName(), CDs: []CD{}}

	allyProducts, err := getAllyProducts(client)
	if err != nil {
		return account, err
	}

	product, found := common.Find(
		allyProducts.CompetitorProdRates.Rates,
		func(product allyCompetitorProduct) bool { return product.ProductCode == "CD" },
	)
	if !found {
		return account, errors.New("product 'CD' was not found")
	}

	bank, found := common.Find(
		product.Banks,
		func(bank allyCompetitorBank) bool { return bank.BankName == "Ally Bank" },
	)
	if !found {
		return account, errors.New("bank 'Ally Bank' was not found")
	}

	for _, term := range bank.Terms {
		if len(term.Rates) <= 0 {
			continue
		}

		cd := CD{
			Term: term.Term,
			Rate: strconv.FormatFloat(term.Rates[0].Apy, 'f', 2, 64),
		}
		account.CDs = append(account.CDs, cd)
	}

	if len(account.CDs) <= 0 {
		return account, errors.New("no CDs found")
	}

	account.sortCdsByTerm()
	return account, nil
}

func (ally *ally) getName() string {
	return "Ally"
}

func (ally *ally) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	account := &SavingsAccount{Name: ally.getName()}

	allyProducts, err := getAllyProducts(client)
	if err != nil {
		return account, err
	}

	product, found := common.Find(
		allyProducts.CompetitorProdRates.Rates,
		func(product allyCompetitorProduct) bool { return product.ProductCode == "OSAV" },
	)
	if !found {
		return account, errors.New("product 'OSAV' was not found")
	}

	bank, found := common.Find(
		product.Banks,
		func(bank allyCompetitorBank) bool { return bank.BankName == "Ally Bank" },
	)
	if !found {
		return account, errors.New("bank 'Ally Bank' was not found")
	}

	if len(bank.Terms) <= 0 {
		return nil, errors.New("no 'terms' found")
	}

	if len(bank.Terms[0].Rates) <= 0 {
		return nil, errors.New("no 'rates' found")
	}

	account.Rate = strconv.FormatFloat(bank.Terms[0].Rates[0].Apy, 'f', 2, 64)
	return account, nil
}
