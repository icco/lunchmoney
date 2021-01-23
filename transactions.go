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

// TransactionFilters are options to pass into the request for transactions.
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

// ToMap converts the filters to a string map to be sent with the request as
// GET parameters.
func (r *TransactionFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &ret); err != nil {
		return nil, err
	}

	return ret, nil
}

// GetTransactions gets all transactions filtered by the filters.
func (c *Client) GetTransactions(ctx context.Context, filters *TransactionFilters) ([]*Transaction, error) {
	validate := validator.New()
	options := map[string]string{}
	if filters != nil {
		if err := validate.Struct(filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, err
		}
		options = maps
	}

	body, err := c.Get(ctx, "/v1/transactions", options)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	resp := &TransactionsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.Transactions, nil
}

// GetTransaction gets a transaction by id.
func (c *Client) GetTransaction(ctx context.Context, id int64, filters *TransactionFilters) (*Transaction, error) {
	validate := validator.New()
	options := map[string]string{}
	if filters != nil {
		// TODO: Turn filters into map.
		if err := validate.Struct(filters); err != nil {
			return nil, err
		}
	}

	body, err := c.Get(ctx, fmt.Sprintf("/v1/transactions/%d", id), options)
	if err != nil {
		return nil, fmt.Errorf("get transaction %d: %w", id, err)
	}

	resp := &Transaction{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
