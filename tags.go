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

// GetTags retrieves all tags from the Lunch Money API.
// It returns a slice of Tag objects containing tag details such as ID, name, and description.
// Returns an error if the request fails or if any tag fails validation.
func (c *Client) GetTags(ctx context.Context) ([]*Tag, error) {
	validate := validator.New()
	body, err := c.Get(ctx, "/v1/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
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
