package lunchmoney

import (
	"context"
	"encoding/json"
)

type User struct {
	UserName        string `json:"user_name"`
	UserEmail       string `json:"user_email"`
	UserID          int    `json:"user_id"`
	AccountID       int    `json:"account_id"`
	BudgetName      string `json:"budget_name"`
	PrimaryCurrency string `json:"primary_currency"`
	APIKeyLabel     string `json:"api_key_label"`
}

func (c *Client) GetUser(ctx context.Context) (*User, error) {
	body, err := c.Get(ctx, "/v1/me", nil)
	if err != nil {
		return nil, err
	}

	resp := &User{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
