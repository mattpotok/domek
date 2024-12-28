package finance

import (
	"strings"
	"sync"

	"github.com/mattpotok/domek/backend/internal/common"
	"github.com/playwright-community/playwright-go"
)

func fetchAllySavings(browserContext playwright.BrowserContext) (*Institution, error) {
	ally := newInstitution("Ally")

	page, err := browserContext.NewPage()
	if err != nil {
		return ally, err
	}

	url := "https://www.ally.com/bank/online-savings-account"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return ally, err
	}

	span := page.Locator("//span[@class='allysf-rates-v1-rate']").First()
	rate, err := span.InnerText()
	if err != nil {
		return ally, err
	}

	if err := page.Close(); err != nil {
		return ally, err
	}

	ally.Savings = strings.TrimSuffix(rate, "%")
	return ally, nil
}

func fetchCapitalOneSavings(browserContext playwright.BrowserContext) (*Institution, error) {
	capitalOne := newInstitution("CapitalOne")

	page, err := browserContext.NewPage()
	if err != nil {
		return capitalOne, err
	}

	url := "https://www.capitalone.com/bank/savings-accounts/online-performance-savings-account/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return capitalOne, err
	}

	banksRateTable := page.Locator("//bank-rates-table").First()
	ratesInline := banksRateTable.Locator("//rates-inline").First()
	rate, err := ratesInline.InnerText()
	if err != nil {
		return capitalOne, err
	}

	if err := page.Close(); err != nil {
		return capitalOne, err
	}

	capitalOne.Savings = strings.TrimSuffix(rate, "%")
	return capitalOne, nil
}

func FetchDiscoverSavings(browserContext playwright.BrowserContext) (*Institution, error) {
	discover := newInstitution("Discover")

	page, err := browserContext.NewPage()
	if err != nil {
		return discover, err
	}

	url := "https://www.discover.com/online-banking/savings-account/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		// Discover website sometimes gets stuck loading for a couple of minutes even though
		// data on page seems to have loaded. Ignore the 'Timeout' error and fetch the present CD rates.
		if !strings.Contains(err.Error(), "Timeout") {
			return discover, err
		}
	}

	div := page.Locator("//div[@analyticsscrolllabel='APY_Rate_Scroll']")
	span := div.Locator("//span[@class='reflectApy']")
	rate, err := span.InnerText()
	if err != nil {
		return discover, err
	}

	if err := page.Close(); err != nil {
		return discover, err
	}

	discover.Savings = strings.TrimSuffix(rate, "%")
	return discover, nil
}

func FetchSavingsAccounts() ([]common.Result[*Institution], error) {
	// TODO remote the `Fatalln`s
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	options := playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)}
	browser, err := pw.Chromium.Launch(options)
	if err != nil {
		return nil, err
	}

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{UserAgent: &userAgent, ExtraHttpHeaders: extraHttpHeaders})
	if err != nil {
		return nil, err
	}

	fetches := [](func(playwright.BrowserContext) (*Institution, error)){
		fetchAllySavings,
		fetchCapitalOneSavings,
		FetchDiscoverSavings,
	}

	institutions := make([]common.Result[*Institution], len(fetches))
	var wg sync.WaitGroup

	for i, fetch := range fetches {
		wg.Add(1)
		go func() {
			defer wg.Done()
			institution, err := fetch(browserContext)
			institutions[i] = common.Result[*Institution]{Value: institution, Error: err}
		}()
	}

	wg.Wait()

	if err := browserContext.Close(); err != nil {
		return nil, err
	}

	if err := browser.Close(); err != nil {
		return nil, err
	}

	if err := pw.Stop(); err != nil {
		return nil, err
	}

	return institutions, nil
}
