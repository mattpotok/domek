package finance

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type vanguardCdRates struct {
	Months3 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"1_3_months"`
	Months6 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"4_6_months"`
	Months9 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"7_9_months"`
	Months12 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"10_12_months"`
	Months18 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"13_18_months"`
	Months24 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"2_years"`
	Months36 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"3_years"`
	Months48 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"4_years"`
	Months60 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"5_years"`
	Months84 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"7_years"`
	Months120 struct {
		Yield         any `json:"yield"`
		EffectiveDate any `json:"effectiveDate"`
	} `json:"10_plus_years"`
}

type vanguard struct{}

func (vanguard *vanguard) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: vanguard.getName(), CDs: []CD{}}

	url := "https://investor.vanguard.com/content/retail/public/us/en/investment-products/cds/jcr:content/root/container/container_copy_16032/container/cdratetable.cdrates.json"
	resp, err := client.Get(url)
	if err != nil {
		return account, nil
	}

	if resp.StatusCode != http.StatusOK {
		return account, errors.New("unexpected status code: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return account, err
	}

	var rates vanguardCdRates
	if err = json.Unmarshal(body, &rates); err != nil {
		return account, err
	}

	t := reflect.TypeOf(rates)
	v := reflect.ValueOf(rates)
	for i := range t.NumField() {
		field := t.Field(i)
		value := v.Field(i)

		if value.Kind() == reflect.Struct {
			yield := value.FieldByName("Yield")
			if yield.IsNil() {
				continue
			}

			rateStr, ok := yield.Interface().(string)
			if !ok {
				continue
			}

			rate, err := strconv.ParseFloat(rateStr, 64)
			if err != nil {
				continue
			}

			termStr, _ := strings.CutPrefix(field.Name, "Months")
			term, err := strconv.Atoi(termStr)
			if err != nil {
				continue
			}

			cd := CD{Term: term, Rate: strconv.FormatFloat(rate, 'f', 2, 64)}
			account.CDs = append(account.CDs, cd)
		}
	}

	if len(account.CDs) <= 0 {
		return account, errors.New("no CDs found")
	}

	account.sortCdsByTerm()
	return account, nil
}

func (vanguard *vanguard) getName() string {
	return "Vanguard"
}

func (vanguard *vanguard) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	account := &SavingsAccount{Name: vanguard.getName()}
	return account, errors.New("not supported")
}
