package finance

import (
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type fidelity struct{}

func (fidelity *fidelity) getCdAccount(client *http.Client) (*CDAccount, error) {
	account := &CDAccount{Name: fidelity.getName(), CDs: []CD{}}

	url := "https://fixedincome.fidelity.com/ftgw/fi/FICDLadder?format=INVEST_PRODUCT_PAGE"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0")

	resp, err := client.Do(req)
	if err != nil {
		return account, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return account, err
	}

	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "gzip" {
		return account, fmt.Errorf("encoding '%s' is not 'gzip'", encoding)
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return account, err
	}
	defer reader.Close()

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return account, err
	}

	termSet := make(map[int]struct{})
	body := doc.Find("div.cd-ladder-bricklets-body").First()
	body.Find("div[class^=cd-ladder-bar-]").Each(func(i int, s *goquery.Selection) {
		if len(s.Children().Nodes) == 0 {
			return
		}

		rate := s.Find("div.cd-ladder-bar-rate").Text()
		if rate == "--" {
			return
		}

		termStr := s.Find("div.cd-ladder-bar-term").Text()
		var term int
		var unit string
		fmt.Sscanf(termStr, "%d %s", &term, &unit)
		if unit == "yr" {
			term *= 12
		}

		if _, found := termSet[term]; found {
			return
		}
		termSet[term] = struct{}{}

		cd := CD{Term: term, Rate: rate}
		account.CDs = append(account.CDs, cd)
	})

	return account, nil
}

func (fidelity *fidelity) getName() string {
	return "Fidelity"
}

func (fidelity *fidelity) getSavingsAccount(client *http.Client) (*SavingsAccount, error) {
	account := &SavingsAccount{Name: fidelity.getName()}
	return account, errors.New("not supported")
}
