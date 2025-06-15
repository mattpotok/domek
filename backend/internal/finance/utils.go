package finance

// The 'sec-ch-ua' header in playwright Chromium by default includes the value "HeadlessChrome";v="X".
// Certain websites use this header value to detect bots and block access.
// Source: https://github.com/microsoft/playwright/issues/27600#issuecomment-2219657708
var extraHttpHeaders map[string]string = map[string]string{
	"sec-ch-ua": `"Not=A?Brand";v="8", "Chromium";v="129"`,
}
var userAgent string = "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0"

/* Available bank rates
 * - Ally       - 3 | 6 | 9 | 12 | 18 | -- | -- | 36 | -- | 60
 * - CapitalOne - - | 6 | 9 | 12 | 18 | 24 | 30 | 36 | 48 | 60
 * - Discover   - 3 | 6 | 9 | 12 | 18 | 24 | 30 | 36 | 48 | 60
 * - Fidelity   - 3 | 6 | 9 | 12 | 18 | 24 | -- | 36 | 48 | 60
 * - Schwab     - 3 | 6 | 9 | 12 | 18 | 24 | -- | -- | -- | --
 */
var terms = []int{3, 6, 9, 12, 18, 24, 30, 36, 48, 60}

var TaxBrackets = [...]float64{10.0, 12.0, 22.0, 24.0, 32.0, 35.0, 37.0}

type CD struct {
	Term int    `json:"term"`
	Rate string `json:"rate"`
}

type institution interface {
	getCdRates() ([]CD, error)
	getName() string
	getSavingsRate() (string, error)
}

type Institution struct {
	Name    string `json:"name"`
	CDs     []CD   `json:"cds"`
	Savings string `json:"savings"`
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

func newInstitution(name string) *Institution {
	cds := make([]CD, len(terms))
	for i, term := range terms {
		cds[i] = CD{Term: term, Rate: ""}
	}

	return &Institution{Name: name, CDs: cds, Savings: ""}
}

func (inst *Institution) setRate(term int, rate string) bool {
	for i, cdRate := range inst.CDs {
		if cdRate.Term == term {
			inst.CDs[i].Rate = rate
			return true
		}
	}

	return false
}
