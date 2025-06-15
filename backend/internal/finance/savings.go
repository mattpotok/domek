package finance

import (
	"sync"

	"github.com/mattpotok/domek/backend/internal/common"
)

// TODO consider folding this in together with cd.go
func GetSavingAccounts() []common.Result[*SavingsAccount] {
	institutions := [](institution){
		&Ally{},
		&CapitalOne{},
		&Discover{},
	}

	accounts := make([]common.Result[*SavingsAccount], len(institutions))
	var wg sync.WaitGroup

	for i, institution := range institutions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO consider passing a client.HttpClient here
			rate, err := institution.getSavingsRate()
			accounts[i] = common.Result[*SavingsAccount]{
				Value: &SavingsAccount{
					Name: institution.getName(),
					Rate: rate,
				},
				Error: err,
			}
		}()
	}

	wg.Wait()

	return accounts
}
