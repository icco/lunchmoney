package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// PlaidAccountsResponse is a list plaid accounts response.
type PlaidAccountsResponse struct {
	PlaidAccounts []*PlaidAccount `json:"plaid_accounts"`
}

// PlaidAccount is a single LM Plaid account.
type PlaidAccount struct {
	ID                int64     `json:"id"`
	DateLinked        string    `json:"date_linked"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Subtype           string    `json:"subtype"`
	Mask              string    `json:"mask"`
	InstitutionName   string    `json:"institution_name"`
	Status            string    `json:"status"`
	LastImport        time.Time `json:"last_import"`
	Balance           string    `json:"balance"`
	ToBase            float64   `json:"to_base"` // the balance converted to the user's primary currency
	Currency          string    `json:"currency"`
	BalanceLastUpdate time.Time `json:"balance_last_update"`
	Limit             int64     `json:"limit"`
}

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (p *PlaidAccount) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(p.Balance, p.Currency)
}

// GetPlaidAccounts gets all plaid accounts filtered by the filters.
func (c *Client) GetPlaidAccounts(ctx context.Context) ([]*PlaidAccount, error) {
	validate := validator.New()
	options := map[string]string{}

	body, err := c.Get(ctx, "/v1/plaid_accounts", options)
	if err != nil {
		return nil, fmt.Errorf("get plaid accounts: %w", err)
	}

	resp := &PlaidAccountsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.PlaidAccounts, nil
}
