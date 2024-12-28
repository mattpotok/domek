package finance

/* TODO
 * - CapitalOne may be failing with timeout errors for some reason
 */

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/mattpotok/domek/backend/internal/common"
	"github.com/playwright-community/playwright-go"
)

func fetchAllyCDRates(ctx playwright.BrowserContext) (*Institution, error) {
	ally := newInstitution("Ally")

	page, err := ctx.NewPage()
	if err != nil {
		return ally, err
	}

	url := "https://www.ally.com/bank/high-yield-cd/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return ally, err
	}

	div := page.Locator("//div[@data-product='HYCD']").First()
	err = div.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30_000),
	})
	if err != nil {
		return ally, err
	}

	anchors, err := div.Locator("a").All()
	if err != nil {
		return ally, err
	}

	for _, anchor := range anchors {
		id, err := anchor.GetAttribute("id")
		if err != nil {
			return ally, err
		}

		parts := strings.Split(id, "-")
		if len(parts) != 2 {
			return ally, fmt.Errorf("expected 'id' to contain one '-'")
		}

		term, err := strconv.Atoi(parts[1])
		if err != nil {
			return ally, err
		}

		rate, err := anchor.GetAttribute("data-rate")
		if err != nil {
			return ally, err
		}

		ally.setRate(term, strings.TrimSuffix(rate, "%"))
	}

	if err := page.Close(); err != nil {
		return ally, err
	}

	return ally, nil
}

func fetchCapitalOneCDRates(context playwright.BrowserContext) (*Institution, error) {
	capitalOne := newInstitution("CapitalOne")

	page, err := context.NewPage()
	if err != nil {
		return capitalOne, err
	}

	url := "https://www.capitalone.com/bank/cds/online-cds/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return capitalOne, err
	}

	rows, err := page.Locator("//div[@id='rates-calculator-table']/table/tbody/tr").All()
	if err != nil {
		return capitalOne, err
	}

	for _, row := range rows {
		entry := row.Locator("rates-inline").First()
		term, err := entry.GetAttribute("term")
		if err != nil {
			return capitalOne, err
		}

		months, err := strconv.Atoi(term[:len(term)-1])
		if err != nil {
			return capitalOne, err
		}

		rate, err := entry.TextContent()
		if err != nil {
			return capitalOne, err
		}

		capitalOne.setRate(months, strings.TrimSuffix(rate, "%"))
	}

	if err := page.Close(); err != nil {
		return capitalOne, err
	}

	return capitalOne, nil
}

func fetchDiscoverCDRates(context playwright.BrowserContext) (*Institution, error) {
	discover := newInstitution("Discover")

	page, err := context.NewPage()
	if err != nil {
		return discover, err
	}
	defer page.Close()

	url := "https://www.discover.com/online-banking/cd/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		// Discover website sometimes gets stuck loading for a couple of minutes even though
		// data on page seems to have loaded. Ignore the 'Timeout' error and fetch the present CD rates.
		if !strings.Contains(err.Error(), "Timeout") {
			return discover, err
		}
	}

	items, err := page.Locator("//div[@id='lSSlideWrapper']/ul/li").All()
	if err != nil {
		return discover, err
	}

	for _, item := range items {
		span := item.Locator("//span[starts-with(@class, 'cd-rate-')]").First()
		class, err := span.GetAttribute("class")
		if err != nil {
			return discover, err
		}

		var months int
		fmt.Sscanf(class, "cd-rate-%d-mon", &months)

		rate, err := span.TextContent()
		if err != nil {
			return discover, err
		}

		discover.setRate(months, strings.TrimSuffix(rate, "%"))
	}

	return discover, nil
}

func fetchFidelityCDRates(context playwright.BrowserContext) (*Institution, error) {
	fidelity := newInstitution("Fidelity")

	page, err := context.NewPage()
	if err != nil {
		return fidelity, err
	}
	defer page.Close()

	url := "https://fixedincome.fidelity.com/ftgw/fi/FICDLadder?format=INVEST_PRODUCT_PAGE"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return fidelity, err
	}

	body := page.Locator("//div[@class='cd-ladder-bricklets-body']").First()
	bars, err := body.Locator("//div[@class='cd-ladder-bars']/div[not(@hidden)]").All()
	if err != nil {
		return fidelity, err
	}

	for _, bar := range bars {
		termDiv := bar.Locator("//div[@class='cd-ladder-bar-term']").First()
		termStr, err := termDiv.TextContent()
		if err != nil {
			return fidelity, err
		}

		var term int
		var unit string
		fmt.Sscanf(termStr, "%d %s", &term, &unit)
		if unit == "yr" {
			term *= 12
		}

		rateDiv := bar.Locator("//div[@class='cd-ladder-bar-rate']").First()
		rate, err := rateDiv.TextContent()
		if err != nil {
			return fidelity, err
		} else if rate == "--" {
			continue
		}

		fidelity.setRate(term, strings.TrimSuffix(rate, "%"))
	}

	return fidelity, nil
}

func fetchSchwabCDRates(context playwright.BrowserContext) (*Institution, error) {
	schwab := newInstitution("Schwab")

	page, err := context.NewPage()
	if err != nil {
		return schwab, err
	}
	defer page.Close()

	url := "https://www.schwab.com/fixed-income/certificates-deposit"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return schwab, err
	}

	section := page.Locator("//section[@class='container bcn-table__container']").First()
	table := section.Locator("//table[contains(@class, 'bcn-table--table-data')]").First()
	headers, err := table.Locator("//thead/tr/th").All()
	if err != nil {
		return schwab, err
	}

	cols, err := table.Locator("//tbody/tr/td").All()
	if err != nil {
		return schwab, err
	}

	if len(headers) != len(cols) {
		return schwab, fmt.Errorf("number of terms don't match APYs")
	}

	// Skip row names
	for i := 1; i < len(headers); i++ {
		header := headers[i]
		termStr, err := header.TextContent()
		if err != nil {
			return schwab, err
		}

		termStr = strings.TrimSpace(termStr)

		var term int
		var unit string
		fmt.Sscanf(termStr, "%d %s CDs", &term, &unit)
		if unit == "Year" {
			term *= 12
		}

		col := cols[i]
		rateStr, err := col.TextContent()
		if err != nil {
			return schwab, err
		}

		// A term may not have any available CDs
		rateStr = strings.TrimSpace(rateStr)
		if strings.HasPrefix(rateStr, "--") {
			continue
		}

		var rate string
		fmt.Sscanf(rateStr, "%s APY", &rate)

		schwab.setRate(term, strings.TrimSuffix(rate, "%"))
	}

	return schwab, nil
}

func FetchCDRates() ([]common.Result[*Institution], error) {
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
		fetchAllyCDRates,
		fetchCapitalOneCDRates,
		fetchDiscoverCDRates,
		fetchFidelityCDRates,
		fetchSchwabCDRates,
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

	if err = browser.Close(); err != nil {
		return nil, err
	}

	if err = pw.Stop(); err != nil {
		return nil, err
	}

	return institutions, err
}
