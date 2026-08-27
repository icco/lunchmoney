package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
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

// ParsedAmount converts the recurring expense's amount and currency into a money.Money object.
// This provides a convenient way to work with the expense amount using the go-money library's
// currency handling capabilities. Returns an error if the amount cannot be parsed.
func (r *RecurringExpense) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(r.Amount, r.Currency)
}

// RecurringExpenseFilters are options to pass to the request.
type RecurringExpenseFilters struct {
	StartDate       string `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
	DebitAsNegative bool   `json:"debit_as_negative"`
}

// ToMap converts the recurring expense filters to a string map to be sent with
// the request as GET parameters. An empty StartDate is omitted.
//
// The map is built field by field rather than marshaled through JSON: a
// map[string]string cannot hold the JSON bool that DebitAsNegative encodes to,
// so the round trip failed for every input, including the zero value.
func (r *RecurringExpenseFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}

	if r.StartDate != "" {
		ret["start_date"] = r.StartDate
	}

	ret["debit_as_negative"] = strconv.FormatBool(r.DebitAsNegative)

	return ret, nil
}

// GetRecurringExpenses retrieves all recurring expenses from the Lunch Money API based on the provided filters.
// It returns a slice of RecurringExpense objects or an error if the request fails.
// The filters parameter can be used to specify date ranges and other criteria.
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
