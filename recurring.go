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

// RecurringItemsResponse is the data struct we get back from a get request.
type RecurringItemsResponse struct {
	RecurringItems []*RecurringItem `json:"recurring_items"`
}

// RecurringItem is a transaction scheduled to happen repeatedly. v1 called
// these recurring expenses and kept the criteria flat.
type RecurringItem struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	// Status is either "suggested" or "reviewed". Only reviewed items are
	// applied to matching transactions.
	Status              string             `json:"status"`
	TransactionCriteria RecurringCriteria  `json:"transaction_criteria"`
	Overrides           RecurringOverrides `json:"overrides"`
	Matches             *RecurringMatches  `json:"matches"`
	CreatedBy           int64              `json:"created_by"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	// Source is one of manual, transaction or system, and is empty for some
	// older recurring items.
	Source string `json:"source"`
}

// RecurringCriteria is the set of properties used to identify transactions that
// match a recurring item.
type RecurringCriteria struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	// Granularity is one of day, week, month or year, and combines with
	// Quantity and AnchorDate to describe the cadence.
	Granularity     string  `json:"granularity"`
	Quantity        int64   `json:"quantity"`
	AnchorDate      string  `json:"anchor_date"`
	Payee           string  `json:"payee"`
	Amount          string  `json:"amount"`
	ToBase          float64 `json:"to_base"`
	Currency        string  `json:"currency"`
	PlaidAccountID  *int64  `json:"plaid_account_id"`
	ManualAccountID *int64  `json:"manual_account_id"`
}

// RecurringOverrides are the values applied to transactions that match a
// recurring item.
type RecurringOverrides struct {
	Payee      string `json:"payee"`
	Notes      string `json:"notes"`
	CategoryID *int64 `json:"category_id"`
}

// RecurringMatches describes the expected, found and missing transactions for
// the requested date range. It is nil for items with a "suggested" status.
type RecurringMatches struct {
	RequestStartDate        string           `json:"request_start_date"`
	RequestEndDate          string           `json:"request_end_date"`
	ExpectedOccurrenceDates []string         `json:"expected_occurrence_dates"`
	FoundTransactions       []RecurringMatch `json:"found_transactions"`
	MissingTransactionDates []string         `json:"missing_transaction_dates"`
}

// RecurringMatch is a single transaction that matched a recurring item.
type RecurringMatch struct {
	Date          string `json:"date"`
	TransactionID int64  `json:"transaction_id"`
}

// ParsedAmount converts the item's expected amount and currency into a money.Money.
func (r *RecurringItem) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(r.TransactionCriteria.Amount, r.TransactionCriteria.Currency)
}

// RecurringItemFilters are options for the request. StartDate and EndDate must
// be given together.
type RecurringItemFilters struct {
	StartDate        string `validate:"required_with=EndDate,omitempty,datetime=2006-01-02"`
	EndDate          string `validate:"required_with=StartDate,omitempty,datetime=2006-01-02"`
	IncludeSuggested *bool
}

// ToMap converts the recurring item filters to a string map to be sent with
// the request as GET parameters. Unset fields are omitted.
func (r *RecurringItemFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}

	if r.StartDate != "" {
		ret["start_date"] = r.StartDate
	}

	if r.EndDate != "" {
		ret["end_date"] = r.EndDate
	}

	if r.IncludeSuggested != nil {
		ret["include_suggested"] = strconv.FormatBool(*r.IncludeSuggested)
	}

	return ret, nil
}

// GetRecurringItems retrieves recurring items matching filters.
func (c *Client) GetRecurringItems(ctx context.Context, filters *RecurringItemFilters) ([]*RecurringItem, error) {
	options := map[string]string{}
	if filters != nil {
		validate := validator.New(validator.WithRequiredStructEnabled())
		if err := validate.StructCtx(ctx, filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, fmt.Errorf("convert filters to map: %w", err)
		}
		options = maps
	}

	body, err := c.Get(ctx, "/recurring_items", options)
	if err != nil {
		return nil, fmt.Errorf("get recurring items: %w", err)
	}

	resp := &RecurringItemsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.RecurringItems, nil
}

// GetRecurringItem retrieves a single recurring item by its ID. The filters
// control the date range used to populate the item's Matches.
func (c *Client) GetRecurringItem(ctx context.Context, id int64, filters *RecurringItemFilters) (*RecurringItem, error) {
	options := map[string]string{}
	if filters != nil {
		validate := validator.New(validator.WithRequiredStructEnabled())
		if err := validate.StructCtx(ctx, filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, fmt.Errorf("convert filters to map: %w", err)
		}
		options = maps
	}

	body, err := c.Get(ctx, fmt.Sprintf("/recurring_items/%d", id), options)
	if err != nil {
		return nil, fmt.Errorf("get recurring item %d: %w", id, err)
	}

	resp := &RecurringItem{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
