// Package insider should runs every day
// on 8:00 AM MSK (00:00 in finviz.com)
// to get the latest data for sell and buy transactions.
package insider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type TransactionType string

const (
	// Nov 21
	insiderDateFormat = "Jan 2 2006"

	// Nov 23 09:17 PM
	insiderSECDateFormat = "Jan 2 3:04 PM 2006"

	Buy  TransactionType = "Buy"
	Sale TransactionType = "Sale"
)

type Browser struct {
	store               Storer
	buyTransactionsURL  string
	sellTransactionsURL string
}

type Storer interface {
	InsertTransactions(context.Context, Transactions) error
}

func New(store Storer) *Browser {
	return &Browser{
		store:               store,
		buyTransactionsURL:  "https://finviz.com/insidertrading.ashx?tc=1",
		sellTransactionsURL: "https://finviz.com/insidertrading.ashx?tc=2",
	}
}

// LastDayTransaction parses all sell and buy transactions
// and returns only transactions from the last day
func (b *Browser) LastDayTransaction() (Transactions, error) {
	txb, err := b.buyTransactions()
	if err != nil {
		return nil, fmt.Errorf("buyTransactions: %w", err)
	}

	txs, err := b.sellTransactions()
	if err != nil {
		return nil, fmt.Errorf("sellTransactions: %w", err)
	}

	return append(txb.lastDay(), txs.lastDay()...), nil
}

// Save saves all transactions to the storer
func (b *Browser) Save(ctx context.Context, tx Transactions) error {
	return b.store.InsertTransactions(ctx, tx)
}

type Transactions []Transaction

// lastDay returns only transactions from the last day
//
// finviz returns N latest transactions, but we need only
// transactions from the last day
func (t Transactions) lastDay() Transactions {
	var lastDay Transactions

	y, m, d := time.Now().Date()
	from := time.Date(y, m, d, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	to := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	for _, transaction := range t {
		if transaction.SEC.NotificationDate.After(from) && transaction.SEC.NotificationDate.Before(to) {
			lastDay = append(lastDay, transaction)
		}
	}

	return lastDay
}

type Transaction struct {
	Ticker          string          `json:"ticker" db:"ticker"`
	Owner           string          `json:"owner" db:"owner"`
	Relationship    string          `json:"relationship" db:"relationship"`
	TransactionDate time.Time       `json:"transaction_date" db:"transaction_date"`
	Transaction     TransactionType `json:"transaction" db:"transaction_type"`
	Cost            float64         `json:"cost" db:"cost"`
	Shares          int             `json:"shares" db:"shares"`
	Value           int             `json:"value" db:"value"`
	SharesTotal     int             `json:"shares_total" db:"shares_total"`
	SEC
}

type SEC struct {
	NotificationDate time.Time `json:"notification_date" db:"notification_date"`
	URL              string    `json:"url" db:"url"`
}

type TransactionTypeCount struct {
	Transaction      TransactionType `json:"transaction" db:"transaction_type"`
	TransactionCount int             `json:"transaction_count" db:"transaction_count"`
	TotalValue       float64         `json:"total_value" db:"total_value"`
}

type RelationshipCount struct {
	Relationship string `json:"relationship" db:"relationship"`
	TransactionTypeCount
}

type TotalTransaction struct {
	Ticker     string  `json:"ticker" db:"ticker"`
	TotalValue float64 `json:"total_value" db:"total_value"`
}

func (t TotalTransaction) FinvizTicker() string {
	return fmt.Sprintf("<a href='https://finviz.com/quote.ashx?t=%s'>%s</a>", t.Ticker, t.Ticker)
}

type Tickers []string

func (t Tickers) Finviz() string {
	tick := strings.Join(t, ",")
	return fmt.Sprintf("<a href='https://finviz.com/screener.ashx?v=340&t=%s&o=ticker'>Open ALL in Finviz Screener</a>", tick)

}

func TransactionTypeToEnum(s string) TransactionType {
	switch s {
	case string(Buy):
		return Buy
	case string(Sale):
		return Sale
	default:
		return ""
	}
}

// buyTransactions returns the list of all buy transactions
func (b *Browser) buyTransactions() (Transactions, error) {
	return b.parse(b.buyTransactionsURL)
}

// sellTransactions returns the list of all sell transactions
func (b *Browser) sellTransactions() (Transactions, error) {
	return b.parse(b.sellTransactionsURL)
}

func (b *Browser) parse(url string) (Transactions, error) {
	var insider []Transaction

	c := colly.NewCollector()
	c.OnHTML(".styled-table-new > tbody > tr", func(e *colly.HTMLElement) {
		// skip the header
		if e.Index == 0 {
			return
		}

		date, err := time.Parse(insiderDateFormat, addYear(e.ChildText("td:nth-child(4)")))
		if err != nil {
			log.Printf("date: %s", err)
			return
		}

		secDate, err := time.Parse(insiderSECDateFormat, addYear(e.ChildText("td:nth-child(10)")))
		if err != nil {
			log.Printf("secDate: %s", err)
			return
		}

		cost, err := strconv.ParseFloat(e.ChildText("td:nth-child(6)"), 64)
		if err != nil {
			log.Printf("cost: %s", err)
			return
		}

		shares, err := strconv.Atoi(removeComma(e.ChildText("td:nth-child(7)")))
		if err != nil {
			log.Printf("shares: %s", err)
			return
		}

		value, err := strconv.Atoi(removeComma(e.ChildText("td:nth-child(8)")))
		if err != nil {
			log.Printf("value: %s", err)
			return
		}

		sharesTotal, err := strconv.Atoi(removeComma(e.ChildText("td:nth-child(9)")))
		if err != nil {
			log.Printf("sharesTotal: %s", err)
			return
		}

		insider = append(insider, Transaction{
			Ticker:          e.ChildText("td:nth-child(1)"),
			Owner:           e.ChildText("td:nth-child(2)"),
			Relationship:    e.ChildText("td:nth-child(3)"),
			TransactionDate: date,
			Transaction:     TransactionTypeToEnum(e.ChildText("td:nth-child(5)")),
			Cost:            cost,
			Shares:          shares,
			Value:           value,
			SharesTotal:     sharesTotal,
			SEC: SEC{
				NotificationDate: secDate,
				URL:              e.ChildAttr("td:nth-child(10) > a", "href"),
			},
		})
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}

	return insider, nil
}

func addYear(date string) string {
	y, _, _ := time.Now().Date()
	return fmt.Sprintf("%s %d", date, y)
}

func removeComma(s string) string {
	return strings.ReplaceAll(s, ",", "")
}
