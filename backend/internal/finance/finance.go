package finance

import (
	"net/http"
	"sync"

	"github.com/mattpotok/domek/backend/internal/common"
)

func GetCdAccounts(client *http.Client) []common.Result[*CDAccount] {
	institutions := [](institution){
		&ally{},
		&capitalOne{},
		&discover{},
		&fidelity{},
		&schwab{},
	}

	accounts := make([]common.Result[*CDAccount], len(institutions))
	var wg sync.WaitGroup

	for i, institution := range institutions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			account, err := institution.getCdAccount(client)
			accounts[i] = common.Result[*CDAccount]{Value: account, Error: err}
		}()
	}

	wg.Wait()

	return accounts
}

func GetSavingAccounts(client *http.Client) []common.Result[*SavingsAccount] {
	institutions := [](institution){
		&ally{},
		&capitalOne{},
		&discover{},
	}

	accounts := make([]common.Result[*SavingsAccount], len(institutions))
	var wg sync.WaitGroup

	for i, institution := range institutions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			account, err := institution.getSavingsAccount(client)
			accounts[i] = common.Result[*SavingsAccount]{Value: account, Error: err}
		}()
	}

	wg.Wait()

	return accounts
}
