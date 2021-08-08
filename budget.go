package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

type BudgetsResponse struct {
	Budgets []*Budget `json:"budgets"`
}

type Budget struct {
	CategoryName      string                 `json:"category_name"`
	CategoryID        int                    `json:"category_id"`
	CategoryGroupName string                 `json:"category_group_name"`
	GroupID           int                    `json:"group_id"`
	IsGroup           bool                   `json:"is_group"`
	IsIncome          bool                   `json:"is_income"`
	ExcludeFromBudget bool                   `json:"exclude_from_budget"`
	ExcludeFromTotals bool                   `json:"exclude_from_totals"`
	Data              map[string]*BudgetData `json:"data"`
	Order             int                    `json:"order"`
}

type BudgetData struct {
	BudgetMonth     string  `json:"budget_month" validate:"datetime=2006-01-02"`
	BudgetToBase    float64 `json:"budget_to_base"`
	BudgetAmount    float64 `json:"budget_amount"`
	BudgetCurrency  string  `json:"budget_currency"`
	SpendingToBase  float64 `json:"spending_to_base"`
	NumTransactions int     `json:"num_transactions"`
}

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (b *BudgetData) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(b.BudgetAmount, b.BudgetCurrency)
}

// GetBudgets returns budgets within a time period.
//
// TODO: Figure out what the arguments are for this call. Undocumented.
func (c *Client) GetBudgets(ctx context.Context) ([]*Budget, error) {
	validate := validator.New()
	body, err := c.Get(ctx, "/v1/budgets", options)
	if err != nil {
		return nil, fmt.Errorf("get budgets: %w", err)
	}

	resp := &BudgetsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.Budgets, nil
}
