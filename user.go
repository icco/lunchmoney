package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
)

// User represents the authenticated user's profile information from the Lunch Money API.
type User struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	AccountID       int64  `json:"account_id"`
	BudgetName      string `json:"budget_name"`
	PrimaryCurrency string `json:"primary_currency"`
	APIKeyLabel     string `json:"api_key_label"`
}

// GetUser retrieves information about the currently authenticated user.
// It returns details such as name, email, ID, and account preferences.
func (c *Client) GetUser(ctx context.Context) (*User, error) {
	body, err := c.Get(ctx, "/me", nil)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	resp := &User{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
