package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/currency"
)

// TransactionsResponse is the response we get from requesting transactions.
type TransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
	Error        string         `json:"error"`
}

// Transaction is a single LM transaction.
type Transaction struct {
	ID             int64  `json:"id"`
	Date           string `json:"date" validate:"datetime=2006-01-02"`
	Payee          string `json:"payee"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	Notes          string `json:"notes"`
	CategoryID     int64  `json:"category_id"`
	RecurringID    int64  `json:"recurring_id"`
	AssetID        int64  `json:"asset_id"`
	PlaidAccountID int64  `json:"plaid_account_id"`
	Status         string `json:"status"`
	IsGroup        bool   `json:"is_group"`
	GroupID        int64  `json:"group_id"`
	ParentID       int64  `json:"parent_id"`
	ExternalID     int64  `json:"external_id"`
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

type TransactionFilters struct {
	TagID           int64  `json:"tag_id"`
	RecurringID     int64  `json:"recurring_id"`
	PlaidAccountID  int64  `json:"plaid_account_id"`
	CategoryID      int64  `json:"category_id"`
	AssetID         int64  `json:"asset_id"`
	Offset          int64  `json:"offset"`
	Limit           int64  `json:"limit"`
	StartDate       string `json:"start_date" validate:"datetime=2006-01-02"`
	EndDate         string `json:"end_date" validate:"datetime=2006-01-02"`
	DebitAsNegative bool   `json:"debit_as_negative"`
}

// GetTransactions gets all transactions filtered by the filters.
func (c *Client) GetTransactions(ctx context.Context, filters *TransactionFilters) ([]*Transaction, error) {
	validate := validator.New()
	options := map[string]string{}
	if filters != nil {
		// TODO: Turn filters into map.
		if err := validate.Struct(filters); err != nil {
			return nil, err
		}
	}

	body, err := c.Get(ctx, "/v1/transactions", options)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	resp := &TransactionsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("bad request: %q", resp.Error)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.Transactions, nil
}
