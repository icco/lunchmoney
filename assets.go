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
	Balance         string    `json:"balance"`
	BalanceAsOf     time.Time `json:"balance_as_of"`
	ToBase          float64   `json:"to_base"` // the balance converted to the user's primary currency
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	InstitutionName string    `json:"institution_name"`
	CreatedAt       time.Time `json:"created_at"`
}

// ParsedAmount turns the currency from lunchmoney into a Go currency.
func (a *Asset) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(a.Balance, a.Currency)
}

// GetAssets gets all assets filtered by the filters.
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

// UpdateAsset updates a singl LM asset.
type UpdateAsset struct {
	TypeName             *string `json:"type_name,omitempty"`
	SubtypeName          *string `json:"subtype_name,omitempty"`
	Name                 *string `json:"name,omitempty"`
	Balance              *string `json:"balance,omitempty"`
	BalanceAsOf          *string `json:"balance_as_of,omitempty"`
	Currency             *string `json:"currency,omitempty"`
	InstitutionName      *string `json:"institution_name,omitempty"`
	ClosedOn             *string `json:"closed_on,omitempty"`
	ExcludedTransactions *bool   `json:"excluded_transactions,omitempty"`
}

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
