package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// AssetsResponse is a response to an asset lookup.
type AssetsResponse struct {
	Assets []*Asset `json:"assets"`
}

// Asset is a single LM asset.
type Asset struct {
	ID              int64     `json:"id"`
	TypeName        string    `json:"type_name"`
	SubtypeName     string    `json:"subtype_name"`
	Name            string    `json:"name"`
	DisplayName     string    `json:"display_name"`
	Balance         string    `json:"balance"`
	BalanceAsOf     time.Time `json:"balance_as_of"`
	ToBase          float64   `json:"to_base"` // the balance converted to the user's primary currency
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	InstitutionName string    `json:"institution_name"`
	CreatedAt       time.Time `json:"created_at"`
}

// ParsedAmount converts the asset's balance and currency into a money.Money object.
// This provides a convenient way to work with the asset's value using the go-money library's
// currency handling capabilities. Returns an error if the balance cannot be parsed.
func (a *Asset) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(a.Balance, a.Currency)
}

// GetAssets retrieves all assets from the Lunch Money API.
// It returns a slice of Asset objects containing information about each asset,
// including balance, institution, and status details. Returns an error if the request fails.
func (c *Client) GetAssets(ctx context.Context) ([]*Asset, error) {
	validate := validator.New()
	options := map[string]string{}

	body, err := c.Get(ctx, "/v1/assets", options)
	if err != nil {
		return nil, fmt.Errorf("get assets: %w", err)
	}

	resp := &AssetsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if err := validate.Struct(resp); err != nil {
		return nil, err
	}

	return resp.Assets, nil
}

// UpdateAsset contains the fields that can be updated for an existing asset.
// Only non-nil fields will be sent in the update request.
type UpdateAsset struct {
	TypeName             *string `json:"type_name,omitempty"`
	SubtypeName          *string `json:"subtype_name,omitempty"`
	Name                 *string `json:"name,omitempty"`
	DisplayName          *string `json:"display_name,omitempty"`
	Balance              *string `json:"balance,omitempty"`
	BalanceAsOf          *string `json:"balance_as_of,omitempty"`
	Currency             *string `json:"currency,omitempty"`
	InstitutionName      *string `json:"institution_name,omitempty"`
	ClosedOn             *string `json:"closed_on,omitempty"`
	ExcludedTransactions *bool   `json:"excluded_transactions,omitempty"`
}

// UpdateAsset modifies an existing asset with the specified ID using the provided fields.
// It returns the updated asset information or an error if the update fails.
// Only fields that are non-nil in the asset parameter will be updated.
func (c *Client) UpdateAsset(ctx context.Context, id int64, asset *UpdateAsset) (*Asset, error) {
	validate := validator.New()
	if err := validate.Struct(asset); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, fmt.Sprintf("/v1/assets/%d", id), asset)
	if err != nil {
		return nil, fmt.Errorf("put asset %d: %w", id, err)
	}

	resp := &Asset{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
