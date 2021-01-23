package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type RecurringExpensesResponse struct {
	RecurringExpenses []*RecurringExpenses `json:"recurring_expenses"`
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

// GetRecurringExpences gets all recurring expenses filtered by the filters.
func (c *Client) GetRecurringExpenses(ctx context.Context) ([]*Transaction, error) {
	validate := validator.New()
	options := map[string]string{}
	body, err := c.Get(ctx, "/v1/transactions", options)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	resp := &RecurringExpensesResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.Transactions, nil
}
