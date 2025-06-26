package finance

import (
	"net/http"
	"sort"
)

// TODO consider adding US Bank (savings) and Vanguard (CDs)

var TaxBrackets = [...]float64{10.0, 12.0, 22.0, 24.0, 32.0, 35.0, 37.0}

type CD struct {
	Term int    `json:"term"`
	Rate string `json:"rate"`
}

type institution interface {
	getCdAccount(client *http.Client) (*CDAccount, error)
	getName() string
	getSavingsAccount(client *http.Client) (*SavingsAccount, error)
}

// TODO consider allowing tiers
type SavingsAccount struct {
	Name string `json:"name"`
	Rate string `json:"rate"`
}

type CDAccount struct {
	Name string `json:"name"`
	CDs  []CD   `json:"cds"`
}

func (account *CDAccount) sortCdsByTerm() {
	sort.Slice(account.CDs, func(i, j int) bool {
		return account.CDs[i].Term < account.CDs[j].Term
	})
}
