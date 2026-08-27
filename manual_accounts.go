package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// ManualAccountsResponse is a response to a manual account lookup.
type ManualAccountsResponse struct {
	ManualAccounts []*ManualAccount `json:"manual_accounts"`
}

// ManualAccount is a single LM manual account. In v1 of the API these were
// called assets.
type ManualAccount struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	InstitutionName string `json:"institution_name"`
	DisplayName     string `json:"display_name"`
	// Type is one of cash, credit, cryptocurrency, employee compensation,
	// investment, loan, other liability, other asset, real estate or vehicle.
	// The v1 "depository" type is now "cash".
	Type                    string         `json:"type"`
	Subtype                 string         `json:"subtype"`
	Balance                 string         `json:"balance"`
	Currency                string         `json:"currency"`
	ToBase                  float64        `json:"to_base"` // the balance converted to the user's primary currency
	BalanceAsOf             Timestamp      `json:"balance_as_of"`
	Status                  string         `json:"status"`
	ClosedOn                string         `json:"closed_on"`
	ExternalID              string         `json:"external_id"`
	CustomMetadata          map[string]any `json:"custom_metadata"`
	ExcludeFromTransactions bool           `json:"exclude_from_transactions"`
	CreatedByName           string         `json:"created_by_name"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

// ParsedAmount converts the account's balance and currency into a money.Money.
func (a *ManualAccount) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(a.Balance, a.Currency)
}

// GetManualAccounts retrieves all manual accounts.
func (c *Client) GetManualAccounts(ctx context.Context) ([]*ManualAccount, error) {
	body, err := c.Get(ctx, "/manual_accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("get manual accounts: %w", err)
	}

	resp := &ManualAccountsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.ManualAccounts, nil
}

// GetManualAccount retrieves a single manual account by its ID.
func (c *Client) GetManualAccount(ctx context.Context, id int64) (*ManualAccount, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/manual_accounts/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get manual account %d: %w", id, err)
	}

	resp := &ManualAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// UpdateManualAccount contains the fields that can be updated for an existing
// manual account. Only non-nil fields are sent in the update request.
type UpdateManualAccount struct {
	Name                    *string         `json:"name,omitempty" validate:"omitnil,min=1,max=45"`
	InstitutionName         *string         `json:"institution_name,omitempty" validate:"omitnil,min=1,max=50"`
	DisplayName             *string         `json:"display_name,omitempty"`
	Type                    *string         `json:"type,omitempty" validate:"omitnil,oneof='cash' 'credit' 'cryptocurrency' 'employee compensation' 'investment' 'loan' 'other liability' 'other asset' 'real estate' 'vehicle'"`
	Subtype                 *string         `json:"subtype,omitempty" validate:"omitnil,min=1,max=100"`
	Balance                 *string         `json:"balance,omitempty"`
	Currency                *string         `json:"currency,omitempty" validate:"omitnil,len=3"`
	BalanceAsOf             *string         `json:"balance_as_of,omitempty"`
	Status                  *string         `json:"status,omitempty" validate:"omitnil,oneof=active closed"`
	ClosedOn                *string         `json:"closed_on,omitempty"`
	ExternalID              *string         `json:"external_id,omitempty" validate:"omitnil,max=75"`
	CustomMetadata          *map[string]any `json:"custom_metadata,omitempty"`
	ExcludeFromTransactions *bool           `json:"exclude_from_transactions,omitempty"`
}

// UpdateManualAccount updates the account with the given ID and returns it.
func (c *Client) UpdateManualAccount(ctx context.Context, id int64, account *UpdateManualAccount) (*ManualAccount, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, account); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, fmt.Sprintf("/manual_accounts/%d", id), account)
	if err != nil {
		return nil, fmt.Errorf("update manual account %d: %w", id, err)
	}

	resp := &ManualAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
