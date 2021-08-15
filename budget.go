package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

type Budget struct {
	CategoryGroupName string                 `json:"category_group_name,omitempty"`
	CategoryID        int                    `json:"category_id"`
	CategoryName      string                 `json:"category_name"`
	Data              map[string]*BudgetData `json:"data,omitempty" validate:"dive"`
	ExcludeFromBudget bool                   `json:"exclude_from_budget"`
	ExcludeFromTotals bool                   `json:"exclude_from_totals"`
	GroupID           int                    `json:"group_id"`
	HasChildren       bool                   `json:"has_children,omitempty"`
	IsGroup           bool                   `json:"is_group,omitempty"`
	IsIncome          bool                   `json:"is_income"`
	Order             int                    `json:"order"`
	Recurring         Recurring              `json:"recurring,omitempty"`
}

type List struct {
	Payee    string  `json:"payee"`
	Amount   string  `json:"amount"`
	Currency string  `json:"currency"`
	ToBase   float64 `json:"to_base"`
}

type Recurring struct {
	Sum  float64 `json:"sum"`
	List []List  `json:"list"`
}

type BudgetData struct {
	BudgetMonth     string      `json:"budget_month,omitempty" validate:"datetime=2006-01-02"`
	BudgetToBase    float64     `json:"budget_to_base,omitempty"`
	BudgetAmount    json.Number `json:"budget_amount,omitempty"`
	BudgetCurrency  string      `json:"budget_currency,omitempty"`
	SpendingToBase  float64     `json:"spending_to_base,omitempty"`
	NumTransactions int         `json:"num_transactions,omitempty"`
}

// BudgetFilters are options to pass into the request for budget history.
type BudgetFilters struct {
	StartDate string `json:"start_date" validate:"datetime=2006-01-02,required"`
	EndDate   string `json:"end_date" validate:"datetime=2006-01-02,required"`
}

// ToMap converts the filters to a string map to be sent with the request as
// GET parameters.
func (r *BudgetFilters) ToMap() (map[string]string, error) {
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

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (b *BudgetData) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(b.BudgetAmount.String(), b.BudgetCurrency)
}

// GetBudgets returns budgets within a time period.
func (c *Client) GetBudgets(ctx context.Context, filters *BudgetFilters) ([]*Budget, error) {
	validate := validator.New()
	options := map[string]string{}
	if filters != nil {
		if err := validate.StructCtx(ctx, filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, err
		}
		options = maps
	}

	body, err := c.Get(ctx, "/v1/budgets", options)
	if err != nil {
		return nil, fmt.Errorf("get budgets: %w", err)
	}

	var resp []*Budget
	var bodyCopy bytes.Buffer
	tee := io.TeeReader(body, &bodyCopy)
	if err := json.NewDecoder(tee).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	for _, b := range resp {
		// Clean up sometimes bad data returned.
		for k, bd := range b.Data {
			if bd.BudgetMonth == "" {
				bd.BudgetMonth = k
			}
		}

		if err := validate.StructCtx(ctx, b); err != nil {
			switch v := err.(type) {
			case validator.ValidationErrors:
				return nil, fmt.Errorf("validating response: %s", v.Error())
			case *validator.InvalidValidationError:
				return nil, fmt.Errorf("validating response (InvalidValidation): %s", v.Error())
			default:
				return nil, fmt.Errorf("validating response (%T): %w", err, v)
			}
		}
	}

	return resp, nil
}
