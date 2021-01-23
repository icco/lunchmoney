package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	InstitutionName string    `json:"institution_name"`
	CreatedAt       time.Time `json:"created_at"`
}

// GetAssets gets all transactions filtered by the filters.
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
