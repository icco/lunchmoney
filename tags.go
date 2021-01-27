package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// TagsResponse is the response from getting all tags.
type TagsResponse []*Tag

// Tag is a single LM tag.
type Tag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetTags gets all tags filtered by the filters.
func (c *Client) GetTags(ctx context.Context) ([]*Tag, error) {
	validate := validator.New()
	body, err := c.Get(ctx, "/v1/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	resp := &TagsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	ret := []*Tag(*resp)

	for _, t := range ret {
		if err := validate.Struct(t); err != nil {
			return nil, err
		}
	}

	return ret, nil
}
