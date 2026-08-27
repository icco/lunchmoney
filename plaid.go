package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
)

// PlaidAccountsResponse is a list plaid accounts response.
type PlaidAccountsResponse struct {
	PlaidAccounts []*PlaidAccount `json:"plaid_accounts"`
}

// PlaidAccount is a single LM Plaid account.
type PlaidAccount struct {
	ID              int64  `json:"id"`
	PlaidItemID     string `json:"plaid_item_id"`
	DateLinked      string `json:"date_linked"`
	LinkedByName    string `json:"linked_by_name"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	Subtype         string `json:"subtype"`
	Mask            string `json:"mask"`
	InstitutionName string `json:"institution_name"`
	// Status is one of active, inactive, closed, deactivated, not found,
	// not supported, relink, syncing, revoked or error.
	Status                        string     `json:"status"`
	AllowTransactionModifications bool       `json:"allow_transaction_modifications"`
	Limit                         *float64   `json:"limit"`
	Balance                       string     `json:"balance"`
	Currency                      string     `json:"currency"`
	ToBase                        float64    `json:"to_base"` // the balance converted to the user's primary currency
	BalanceLastUpdate             *time.Time `json:"balance_last_update"`
	ImportStartDate               string     `json:"import_start_date"`
	LastImport                    *time.Time `json:"last_import"`
	LastFetch                     *time.Time `json:"last_fetch"`
	PlaidLastSuccessfulUpdate     *time.Time `json:"plaid_last_successful_update"`
}

// ParsedAmount converts the Plaid account balance and currency into a money.Money object.
// This provides a convenient way to work with account balances using the go-money library's
// currency handling capabilities. Returns an error if the balance cannot be parsed.
func (p *PlaidAccount) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(p.Balance, p.Currency)
}

// GetPlaidAccounts retrieves all Plaid-connected accounts from the Lunch Money API.
// It returns a slice of PlaidAccount objects containing information about each account,
// including balance, institution information, and status.
func (c *Client) GetPlaidAccounts(ctx context.Context) ([]*PlaidAccount, error) {
	body, err := c.Get(ctx, "/plaid_accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("get plaid accounts: %w", err)
	}

	resp := &PlaidAccountsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.PlaidAccounts, nil
}

// GetPlaidAccount retrieves a single Plaid-connected account by its ID.
func (c *Client) GetPlaidAccount(ctx context.Context, id int64) (*PlaidAccount, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/plaid_accounts/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get plaid account %d: %w", id, err)
	}

	resp := &PlaidAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
