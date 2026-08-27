package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// Category format options accepted by GetCategories.
const (
	// CategoryFormatNested returns category groups with their members in a
	// Children slice. This is the API default.
	CategoryFormatNested = "nested"

	// CategoryFormatFlattened returns every category at the top level, which
	// is what v1 of the API did.
	CategoryFormatFlattened = "flattened"
)

// CategoriesResponse is the response we get from requesting categories.
type CategoriesResponse struct {
	Categories []*Category `json:"categories"`
}

// Category represents a single Lunch Money category.
// Categories are used to organize transactions and budgets.
// They can be grouped hierarchically and marked as income or excluded from various calculations.
type Category struct {
	ID                int64       `json:"id"`                  // Unique identifier for the category
	Name              string      `json:"name"`                // Display name of the category
	Description       string      `json:"description"`         // Optional description of the category
	IsIncome          bool        `json:"is_income"`           // Whether this category represents income
	ExcludeFromBudget bool        `json:"exclude_from_budget"` // Whether to exclude from budget calculations
	ExcludeFromTotals bool        `json:"exclude_from_totals"` // Whether to exclude from total calculations
	UpdatedAt         time.Time   `json:"updated_at"`          // Last modification timestamp
	CreatedAt         time.Time   `json:"created_at"`          // Creation timestamp
	IsGroup           bool        `json:"is_group"`            // Whether this category is a group
	GroupID           *int64      `json:"group_id"`            // ID of the parent group, or nil
	Children          []*Category `json:"children,omitempty"`  // Members of this group, when requesting the nested format
	Archived          bool        `json:"archived"`            // Whether the category is hidden in the app
	ArchivedAt        *time.Time  `json:"archived_at"`         // When the category was last archived, or nil
	Order             *int64      `json:"order"`               // Position on the categories page, or nil for alphabetical
	Collapsed         bool        `json:"collapsed"`           // Whether the group appears collapsed in the app
}

// CategoryFilters are options to pass into the request for categories.
type CategoryFilters struct {
	// Format is either CategoryFormatNested or CategoryFormatFlattened. An
	// empty value uses the API default, which is nested.
	Format string `validate:"omitempty,oneof=nested flattened"`

	// IsGroup, when set, restricts the response to category groups or to
	// categories that are not groups.
	IsGroup *bool
}

// ToMap converts the category filters to a string map to be sent with the
// request as GET parameters. Unset fields are omitted.
func (r *CategoryFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}

	if r.Format != "" {
		ret["format"] = r.Format
	}

	if r.IsGroup != nil {
		ret["is_group"] = fmt.Sprintf("%t", *r.IsGroup)
	}

	return ret, nil
}

// GetCategories returns all categories associated with the user's account.
//
// Unlike v1, the API defaults to the nested format: category groups come back
// with their members in Children rather than as separate top level entries.
// Pass a filter with Format set to CategoryFormatFlattened for the old shape.
func (c *Client) GetCategories(ctx context.Context, filters *CategoryFilters) ([]*Category, error) {
	options := map[string]string{}
	if filters != nil {
		validate := validator.New(validator.WithRequiredStructEnabled())
		if err := validate.StructCtx(ctx, filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, fmt.Errorf("convert filters to map: %w", err)
		}
		options = maps
	}

	body, err := c.Get(ctx, "/categories", options)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	resp := &CategoriesResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.Categories, nil
}

// GetCategory retrieves a single category by its ID.
// It returns detailed information about the category including its metadata,
// group relationships, and various settings.
func (c *Client) GetCategory(ctx context.Context, id int64) (*Category, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/categories/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", id, err)
	}

	resp := &Category{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
