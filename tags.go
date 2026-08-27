package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TagsResponse is the response from getting all tags.
type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

// Tag is a single LM tag.
type Tag struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	TextColor       string     `json:"text_color"`
	BackgroundColor string     `json:"background_color"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Archived        bool       `json:"archived"`
	ArchivedAt      *time.Time `json:"archived_at"`
}

// GetTags retrieves all tags, ordered alphabetically.
func (c *Client) GetTags(ctx context.Context) ([]*Tag, error) {
	body, err := c.Get(ctx, "/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	resp := &TagsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.Tags, nil
}

// GetTag retrieves a single tag by its ID.
func (c *Client) GetTag(ctx context.Context, id int64) (*Tag, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/tags/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get tag %d: %w", id, err)
	}

	resp := &Tag{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
