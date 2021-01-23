package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/text/currency"
)

type RecurringExpensesResponse struct {
	RecurringExpenses []*RecurringExpense `json:"recurring_expenses"`
}

type RecurringExpense struct {
	ID             int64     `json:"id"`
	StartDate      string    `json:"start_date" validate:"datetime=2006-01-02"`
	EndDate        string    `json:"end_date" validate:"datetime=2006-01-02"`
	Cadence        string    `json:"cadence"`
	Payee          string    `json:"payee"`
	Amount         string    `json:"amount"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"created_at"`
	Description    string    `json:"description"`
	BillingDate    string    `json:"billing_date"`
	Type           string    `json:"type"`
	OriginalName   string    `json:"original_name"`
	Source         string    `json:"source"`
	PlaidAccountID int64     `json:"plaid_account_id"`
	AssetID        int64     `json:"asset_id"`
	TransactionID  int64     `json:"transaction_id"`
}

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (r *RecurringExpense) ParsedAmount() (currency.Amount, error) {
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

type RecurringExpenseFilters struct {
	StartDate       string `json:"start_date" validate:"datetime=2006-01-02"`
	DebitAsNegative bool   `json:"debit_as_negative"`
}

func (r *RecurringExpenseFilters) ToMap() (map[string]string, error) {
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

// GetRecurringExpences gets all recurring expenses filtered by the filters.
func (c *Client) GetRecurringExpenses(ctx context.Context, filters *RecurringExpenseFilters) ([]*RecurringExpense, error) {
	validate := validator.New()
	options := map[string]string{}
	if r != nil {
		if err := validate.Struct(filters); err != nil {
			return nil, err
		}

		maps, err := r.tomap()
		if err != nil {
			return nil, err
		}
		options = maps
	}

	body, err := c.Get(ctx, "/v1/recurring_expenses", options)
	if err != nil {
		return nil, fmt.Errorf("get recurring expenses: %w", err)
	}

	resp := &RecurringExpensesResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.RecurringExpenses, nil
}
