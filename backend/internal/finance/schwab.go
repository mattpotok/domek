package finance

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type schwab struct{}

func (schwab *schwab) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: schwab.getName(), CDs: []CD{}}

	url := "https://www.schwab.com/fixed-income/certificates-deposit"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return account, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0")

	resp, err := client.Do(req)
	if err != nil {
		return account, err
	}

	if resp.StatusCode != http.StatusOK {
		return account, errors.New("unexpected status code: " + resp.Status)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return account, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return account, err
	}

	section := doc.Find("section.container.bcn-table__container").First()
	table := section.Find("table.bcn-table--table-data").First()

	headers := table.Find("th").Slice(1, goquery.ToEnd)
	terms := goquery.Map(headers, func(i int, s *goquery.Selection) int {
		text := strings.TrimSpace(s.Text())

		var term int
		var unit string
		fmt.Sscanf(text, "%d %s CDs", &term, &unit)
		if unit == "Year" {
			term *= 12
		}

		return term
	})

	data := table.Find("td").Slice(1, goquery.ToEnd)
	rates := goquery.Map(data, func(i int, s *goquery.Selection) string {
		text := strings.TrimSpace(s.Text())
		if strings.HasPrefix(text, "--") {
			return "N/A"
		}

		end := strings.Index(text, "%")
		return text[0:end]
	})

	if len(terms) != len(rates) {
		log.Fatal(errors.New("number of terms and APYs don't match"))
	}

	for i := 0; i < len(terms); i++ {
		cd := CD{Term: terms[i], Rate: rates[i]}
		account.CDs = append(account.CDs, cd)
	}

	if len(account.CDs) <= 0 {
		return account, errors.New("no CDs found")
	}

	account.sortCdsByTerm()
	return account, nil
}

func (schwab *schwab) getName() string {
	return "Schwab"
}

func (schwab *schwab) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	account := &SavingsAccount{Name: schwab.getName()}
	return account, errors.New("not supported")
}
