package lunchmoney

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/text/currency"
)

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	Error        string        `json:"error"`
}

type Transaction struct {
	ID             int64     `json:"id"`
	Date           time.Time `json:"date"`
	Payee          string    `json:"payee"`
	Amount         string    `json:"amount"`
	Currency       string    `json:"currency"`
	Notes          string    `json:"notes"`
	CategoryID     int64     `json:"category_id"`
	RecurringID    int64     `json:"recurring_id"`
	AssetID        int64     `json:"asset_id"`
	PlaidAccountID int64     `json:"plaid_account_id"`
	Status         string    `json:"status"`
	IsGroup        bool      `json:"is_group"`
	GroupID        int64     `json:"group_id"`
	ParentID       int64     `json:"parent_id"`
	ExternalID     int64     `json:"external_id"`
}

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (t *Transaction) ParsedAmount() (currency.Amount, error) {
	cur, err := currency.ParseISO(t.Currency)
	if err != nil {
		return currency.Amount{}, fmt.Errorf("%q is not valid currency: %w", t.Currency, err)
	}

	f, err := strconv.ParseFloat(t.Amount, 64)
	if err != nil {
		return currency.Amount{}, fmt.Errorf("%q is not valid float: %w", t.Amount, err)
	}

	return cur.Amount(f), nil
}
