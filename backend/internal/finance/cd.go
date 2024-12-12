package finance

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

type CD struct {
	Term int    `json:"term"`
	Rate string `json:"rate"`
}

type Institution struct {
	Name string `json:"name"`
	CDs  []CD   `json:"cds"`
}

type channelData struct {
	Index       int
	Institution *Institution
	Error       error
}

const USER_AGENT string = "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0"

func (inst *Institution) SetRate(term int, rate string) bool {
	for i, cdRate := range inst.CDs {
		if cdRate.Term == term {
			inst.CDs[i].Rate = rate
			return true
		}
	}

	return false
}

func NewInstitution(name string) *Institution {
	terms := []int{3, 6, 9, 12, 18, 24, 30, 36, 48, 60}

	cds := make([]CD, len(terms))
	for i, term := range terms {
		cds[i] = CD{Term: term, Rate: "N/A"}
	}

	return &Institution{Name: name, CDs: cds}
}

func fetchAllyCDRates(context playwright.BrowserContext) (*Institution, error) {
	ally := NewInstitution("Ally")

	page, err := context.NewPage()
	if err != nil {
		return ally, err
	}
	defer page.Close()

	url := "https://www.ally.com/bank/high-yield-cd/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return ally, err
	}

	// FIXME rename this
	locator := page.Locator("//div[@data-product='HYCD']").First()
	err = locator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30_000),
	})
	if err != nil {
		return ally, err
	}

	anchors, err := locator.Locator("a").All()
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
			return nil, fmt.Errorf("expected 'id' to contain one '-'")
		}

		term, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}

		rate, err := anchor.GetAttribute("data-rate")
		if err != nil {
			return ally, err
		}

		ally.SetRate(term, rate+"%")
	}

	return ally, nil
}

func fetchCapitalOneCDRates(context playwright.BrowserContext) (*Institution, error) {
	capitalOne := NewInstitution("CapitalOne")

	page, err := context.NewPage()
	if err != nil {
		return capitalOne, err
	}
	defer page.Close()

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

		capitalOne.SetRate(months, rate)
	}

	return capitalOne, nil
}

func fetchDiscoverCDRates(context playwright.BrowserContext) (*Institution, error) {
	discover := NewInstitution("Discover")

	page, err := context.NewPage()
	if err != nil {
		return discover, err
	}
	defer page.Close()

	url := "https://www.discover.com/online-banking/cd/"
	_, err = page.Goto(url, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateLoad})
	if err != nil {
		return discover, err
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

		discover.SetRate(months, rate+"%")
	}

	return discover, nil
}

func fetchFidelityCDRates(context playwright.BrowserContext) (*Institution, error) {
	fidelity := NewInstitution("Fidelity")

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
		}

		fidelity.SetRate(term, rate+"%")
	}

	return fidelity, nil
}

func fetchSchwabCDRates(context playwright.BrowserContext) (*Institution, error) {
	schwab := NewInstitution("Schwab")

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

		schwab.SetRate(term, rate)
	}

	return schwab, nil
}

/* Available bank rates
 * - Ally       - 3 | 6 | 9 | 12 | 18 | -- | -- | 36 | -- | 60
 * - CapitalOne - - | 6 | 9 | 12 | 18 | 24 | 30 | 36 | 48 | 60
 * - Discover   - 3 | 6 | 9 | 12 | 18 | 24 | 30 | 36 | 48 | 60
 * - Fidelity   - 3 | 6 | 9 | 12 | 18 | 24 | -- | 36 | 48 | 60
 * - Schwab     - 3 | 6 | 9 | 12 | 18 | 24 | -- | -- | -- | --
 */
func FetchCDRates() ([]Institution, error) {
	// TODO remote the `Fatalln`s
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
	if err != nil {
		log.Fatalln(err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalln(err)
	}

	options := playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)}
	browser, err := pw.Chromium.Launch(options)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO document fix https://github.com/microsoft/playwright/issues/27600#issuecomment-2219657708
	//
	userAgent := "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0"
	extraHttpHeaders := map[string]string{
		"sec-ch-ua": `"Not=A?Brand";v="8", "Chromium";v="129"`,
	}
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{UserAgent: &userAgent, ExtraHttpHeaders: extraHttpHeaders})
	if err != nil {
		log.Fatalln(err)
	}

	fetches := [](func(playwright.BrowserContext) (*Institution, error)){
		fetchAllyCDRates,
		fetchCapitalOneCDRates,
		fetchDiscoverCDRates,
		fetchFidelityCDRates,
		fetchSchwabCDRates,
	}

	channel := make(chan channelData)
	var wg sync.WaitGroup

	for i, fetch := range fetches {
		wg.Add(1)

		go func() {
			defer wg.Done()
			institution, err := fetch(context)
			channel <- channelData{Index: i, Institution: institution, Error: err}
		}()
	}

	go func() {
		wg.Wait()
		close(channel)
	}()

	institutions := make([]Institution, len(fetches))
	for data := range channel {
		institutions[data.Index] = *data.Institution
		if data.Error != nil {
			// TODO combine the errors into one here
			fmt.Println(err)
		}
	}

	if err = browser.Close(); err != nil {
		return institutions, err
	}

	if err = pw.Stop(); err != nil {
		return institutions, err
	}

	return institutions, err
}
