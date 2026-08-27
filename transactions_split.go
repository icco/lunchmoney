package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// SplitTransactionRequest splits one transaction into child transactions. The
// children's amounts must sum to the parent's amount.
type SplitTransactionRequest struct {
	ChildTransactions []SplitTransactionChild `json:"child_transactions" validate:"min=2,max=500,dive"`
}

// SplitTransactionChild is one child of a split. Only Amount is required;
// every other field is inherited from the parent when left unset.
type SplitTransactionChild struct {
	Amount     string  `json:"amount" validate:"required"`
	Payee      string  `json:"payee,omitempty" validate:"max=140"`
	Date       string  `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	CategoryID *int64  `json:"category_id,omitempty"`
	TagIDs     []int64 `json:"tag_ids,omitempty"`
	Notes      string  `json:"notes,omitempty"`
}

// SplitTransaction splits a transaction and returns the split parent with its
// new children populated.
func (c *Client) SplitTransaction(ctx context.Context, id int64, req SplitTransactionRequest) (*Transaction, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, req); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, fmt.Sprintf("/transactions/split/%d", id), req)
	if err != nil {
		return nil, fmt.Errorf("split transaction %d: %w", id, err)
	}

	resp := &Transaction{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// UnsplitTransaction deletes the children of a split transaction and restores
// the parent. In v1 this was POST /transactions/unsplit.
func (c *Client) UnsplitTransaction(ctx context.Context, id int64) error {
	// The API answers with 204 and no body.
	if _, err := c.Delete(ctx, fmt.Sprintf("/transactions/split/%d", id), nil); err != nil {
		return fmt.Errorf("unsplit transaction %d: %w", id, err)
	}

	return nil
}

// GroupTransactionsRequest groups transactions into a new one. Date and Payee
// are sent even when empty; the API requires both keys.
type GroupTransactionsRequest struct {
	IDs        []int64 `json:"ids" validate:"min=2,max=500"`
	Date       string  `json:"date" validate:"datetime=2006-01-02"`
	Payee      string  `json:"payee" validate:"max=140"`
	CategoryID *int64  `json:"category_id,omitempty"`
	Notes      string  `json:"notes,omitempty"`
	Status     string  `json:"status,omitempty" validate:"omitempty,oneof=reviewed unreviewed"`
	TagIDs     []int64 `json:"tag_ids,omitempty"`
}

// GroupTransactions groups transactions and returns the new group parent with
// its children populated.
func (c *Client) GroupTransactions(ctx context.Context, req GroupTransactionsRequest) (*Transaction, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, req); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, "/transactions/group", req)
	if err != nil {
		return nil, fmt.Errorf("group transactions: %w", err)
	}

	resp := &Transaction{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// UngroupTransactions deletes a transaction group, leaving its members as
// ungrouped transactions.
func (c *Client) UngroupTransactions(ctx context.Context, id int64) error {
	// The API answers with 204 and no body.
	if _, err := c.Delete(ctx, fmt.Sprintf("/transactions/group/%d", id), nil); err != nil {
		return fmt.Errorf("ungroup transactions %d: %w", id, err)
	}

	return nil
}
