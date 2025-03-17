package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
)

// CategoriesResponse is the response we get from requesting categories.
// It contains a list of categories and an optional error message.
type CategoriesResponse struct {
	Categories []*Category `json:"categories"`
	Error      string      `json:"error"`
}

// Category represents a single Lunch Money category.
// Categories are used to organize transactions and budgets.
// They can be grouped hierarchically and marked as income or excluded from various calculations.
type Category struct {
	ID                int64     `json:"id"`                  // Unique identifier for the category
	Name              string    `json:"name"`                // Display name of the category
	Description       string    `json:"description"`         // Optional description of the category
	IsIncome          bool      `json:"is_income"`           // Whether this category represents income
	ExcludeFromBudget bool      `json:"exclude_from_budget"` // Whether to exclude from budget calculations
	ExcludeFromTotals bool      `json:"exclude_from_totals"` // Whether to exclude from total calculations
	UpdatedAt         time.Time `json:"updated_at"`          // Last modification timestamp
	CreatedAt         time.Time `json:"created_at"`          // Creation timestamp
	IsGroup           bool      `json:"is_group"`            // Whether this category is a group
	GroupID           int64     `json:"group_id"`            // ID of the parent group, if any
}

// GetCategories returns a flattened list of all categories in alphabetical
// order associated with the user's account. This includes both regular categories
// and category groups. The returned categories include metadata such as creation time,
// group relationships, and budget exclusion settings.
//
// The context can be used to control the request lifecycle.
// Returns an error if the API request fails or if the response cannot be validated.
func (c *Client) GetCategories(ctx context.Context) ([]*Category, error) {
	validate := validator.New()
	options := map[string]string{}
	body, err := c.Get(ctx, "/v1/categories", options)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	var resp *CategoriesResponse
	var bodyCopy bytes.Buffer
	tee := io.TeeReader(body, &bodyCopy)
	if err := json.NewDecoder(tee).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	for _, b := range resp.Categories {
		if err := validate.StructCtx(ctx, b); err != nil {
			var validationErrors validator.ValidationErrors
			var invalidValidationError *validator.InvalidValidationError

			switch {
			case errors.As(err, &validationErrors):
				return nil, fmt.Errorf("validating response: %s", validationErrors.Error())
			case errors.As(err, &invalidValidationError):
				return nil, fmt.Errorf("validating response (InvalidValidation): %s", invalidValidationError.Error())
			default:
				return nil, fmt.Errorf("validating response (%T): %w", err, err)
			}
		}
	}
	return resp.Categories, nil
}

// GetCategory retrieves a single category by its ID.
// It returns detailed information about the category including its metadata,
// group relationships, and various settings.
//
// Parameters:
//   - ctx: Context for controlling the request lifecycle
//   - id: The unique identifier of the category to retrieve
//
// Returns the category details or an error if the request fails or
// the response cannot be validated.
func (c *Client) GetCategory(ctx context.Context, id int64) (*Category, error) {
	options := map[string]string{}
	body, err := c.Get(ctx, fmt.Sprintf("/v1/categories/%d", id), options)
	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	var resp *Category
	var bodyCopy bytes.Buffer
	tee := io.TeeReader(body, &bodyCopy)
	if err := json.NewDecoder(tee).Decode(&resp); err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	validate := validator.New()
	if err := validate.StructCtx(ctx, resp); err != nil {
		var validationErrors validator.ValidationErrors
		var invalidValidationError *validator.InvalidValidationError

		switch {
		case errors.As(err, &validationErrors):
			return nil, fmt.Errorf("validating response: %s", validationErrors.Error())
		case errors.As(err, &invalidValidationError):
			return nil, fmt.Errorf("validating response (InvalidValidation): %s", invalidValidationError.Error())
		default:
			return nil, fmt.Errorf("validating response (%T): %w", err, err)
		}
	}

	return resp, nil
}
