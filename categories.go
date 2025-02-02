package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
)

// CategoriesResponse is the response we get from requesting categories.
type CategoriesResponse struct {
	Categories []*Category `json:"categories"`
	Error      string      `json:"error"`
}

// Category is a single LM category.
type Category struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	IsIncome          bool      `json:"is_income"`
	ExcludeFromBudget bool      `json:"exclude_from_budget"`
	ExcludeFromTotals bool      `json:"exclude_from_totals"`
	UpdatedAt         time.Time `json:"updated_at"`
	CreatedAt         time.Time `json:"created_at"`
	IsGroup           bool      `json:"is_group"`
	GroupID           int64     `json:"group_id"`
}

// GetCategories returns categories 
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
			switch v := err.(type) {
			case validator.ValidationErrors:
				return nil, fmt.Errorf("validating response: %s", v.Error())
			case *validator.InvalidValidationError:
				return nil, fmt.Errorf("validating response (InvalidValidation): %s", v.Error())
			default:
				return nil, fmt.Errorf("validating response (%T): %w", err, v)
			}
		}
	}
	return resp.Categories, nil
}

func (c *Client) GetCategory( ctx context.Context, id int64) (*Category, error)  {
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
		switch v := err.(type) {
		case validator.ValidationErrors:
			return nil, fmt.Errorf("validating response: %s", v.Error())
		case *validator.InvalidValidationError:
			return nil, fmt.Errorf("validating response (InvalidValidation): %s", v.Error())
		default:
			return nil, fmt.Errorf("validating response (%T): %w", err, v)
		}
	}

	return resp, nil
}