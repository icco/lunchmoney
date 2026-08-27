package lunchmoney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
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

// CreateTag is a tag to create. Name is required and has to be unique.
type CreateTag struct {
	Name            string `json:"name" validate:"required,min=1,max=100"`
	Description     string `json:"description,omitempty" validate:"max=200"`
	TextColor       string `json:"text_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	Archived        bool   `json:"archived,omitempty"`
}

// CreateTag creates a new tag and returns it.
func (c *Client) CreateTag(ctx context.Context, tag *CreateTag) (*Tag, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, tag); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, "/tags", tag)
	if err != nil {
		return nil, fmt.Errorf("create tag: %w", err)
	}

	resp := &Tag{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// UpdateTag holds the updatable fields of a tag. Only non-nil fields are sent,
// which also means there is no way to send an explicit null: leaving ArchivedAt
// nil keeps the stored value rather than clearing it.
type UpdateTag struct {
	Name            *string    `json:"name,omitempty" validate:"omitnil,min=1,max=100"`
	Description     *string    `json:"description,omitempty" validate:"omitnil,max=200"`
	TextColor       *string    `json:"text_color,omitempty"`
	BackgroundColor *string    `json:"background_color,omitempty"`
	Archived        *bool      `json:"archived,omitempty"`
	ArchivedAt      *time.Time `json:"archived_at,omitempty"`
}

// UpdateTag modifies the tag with the given ID and returns it.
func (c *Client) UpdateTag(ctx context.Context, id int64, tag *UpdateTag) (*Tag, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, tag); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, fmt.Sprintf("/tags/%d", id), tag)
	if err != nil {
		return nil, fmt.Errorf("update tag %d: %w", id, err)
	}

	resp := &Tag{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// TagDependents counts the records that still reference a tag.
type TagDependents struct {
	Rules        int64 `json:"rules"`
	Transactions int64 `json:"transactions"`
}

// TagInUseError is the 422 DeleteTag gets back when rules or transactions still
// reference the tag. Deleting it anyway takes another call with force set.
type TagInUseError struct {
	TagName    string        `json:"tag_name"`
	Dependents TagDependents `json:"dependents"`

	// Err is the API error the dependents arrived as.
	Err error `json:"-"`
}

func (e *TagInUseError) Error() string {
	return fmt.Sprintf("tag %q is used by %d rules and %d transactions", e.TagName, e.Dependents.Rules, e.Dependents.Transactions)
}

func (e *TagInUseError) Unwrap() error { return e.Err }

// DeleteTag deletes the tag with the given ID. Without force the API refuses to
// delete a tag that rules or transactions still reference, and the returned
// error is a *TagInUseError naming what depends on it.
func (c *Client) DeleteTag(ctx context.Context, id int64, force bool) error {
	options := map[string]string{}
	if force {
		options["force"] = "true"
	}

	if _, err := c.Delete(ctx, fmt.Sprintf("/tags/%d", id), options); err != nil {
		if inUse := tagInUse(err); inUse != nil {
			return fmt.Errorf("delete tag %d: %w", id, inUse)
		}

		return fmt.Errorf("delete tag %d: %w", id, err)
	}

	return nil
}

// tagInUse decodes the dependents out of a 422, or returns nil if err is
// anything else.
func tagInUse(err error) *TagInUseError {
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity {
		return nil
	}

	inUse := &TagInUseError{Err: err}
	if decodeErr := json.Unmarshal(apiErr.RawBody, inUse); decodeErr != nil || inUse.TagName == "" {
		return nil
	}

	return inUse
}
