package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// RecurringExpensesResponse is the data struct we get back from a get request.
type RecurringExpensesResponse struct {
	RecurringExpenses []*RecurringExpense `json:"recurring_expenses"`
}

// RecurringExpense is like a transaction, but one that's scheduled to happen.
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
func (r *RecurringExpense) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(r.Amount, r.Currency)
}

// RecurringExpenseFilters are options to pass to the request.
type RecurringExpenseFilters struct {
	StartDate       string `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
	DebitAsNegative bool   `json:"debit_as_negative"`
}

// ToMap converts the filters to a string map to be sent with the request as
// GET parameters.
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

// GetRecurringExpenses gets all recurring expenses filtered by the filters.
func (c *Client) GetRecurringExpenses(ctx context.Context, filters *RecurringExpenseFilters) ([]*RecurringExpense, error) {
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
