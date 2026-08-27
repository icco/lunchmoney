package lunchmoney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCategories(t *testing.T) {
	groupID := int64(7)

	tests := []struct {
		name        string
		filters     *CategoryFilters
		wantQuery   string
		response    string
		statusCode  int
		wantErr     bool
		errContains string
		want        []*Category
	}{
		{
			name: "successful response",
			response: `{
				"categories": [
					{
						"id": 1,
						"name": "Groceries",
						"description": "Food and household items",
						"is_income": false,
						"exclude_from_budget": false,
						"exclude_from_totals": false,
						"updated_at": "2023-01-01T00:00:00Z",
						"created_at": "2023-01-01T00:00:00Z",
						"is_group": false,
						"group_id": 7,
						"archived": false,
						"archived_at": null,
						"order": null,
						"collapsed": false
					}
				]
			}`,
			statusCode: http.StatusOK,
			want: []*Category{
				{
					ID:                1,
					Name:              "Groceries",
					Description:       "Food and household items",
					IsIncome:          false,
					ExcludeFromBudget: false,
					ExcludeFromTotals: false,
					UpdatedAt:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					CreatedAt:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					IsGroup:           false,
					GroupID:           &groupID,
				},
			},
		},
		{
			name:       "null group id",
			response:   `{"categories": [{"id": 1, "name": "Groceries", "group_id": null}]}`,
			statusCode: http.StatusOK,
			want:       []*Category{{ID: 1, Name: "Groceries"}},
		},
		{
			name:       "flattened format",
			filters:    &CategoryFilters{Format: CategoryFormatFlattened},
			wantQuery:  "format=flattened",
			response:   `{"categories": []}`,
			statusCode: http.StatusOK,
			want:       []*Category{},
		},
		{
			name:        "invalid format is rejected before the request",
			filters:     &CategoryFilters{Format: "sideways"},
			wantErr:     true,
			errContains: "Format",
		},
		{
			name:        "invalid response",
			response:    `{"invalid": "json"`,
			statusCode:  http.StatusOK,
			wantErr:     true,
			errContains: "decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/categories", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.wantQuery, r.URL.RawQuery)
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).GetCategories(context.Background(), tt.filters)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		response    string
		statusCode  int
		wantErr     bool
		errContains string
		want        *Category
	}{
		{
			name: "successful response",
			id:   1,
			response: `{
				"id": 1,
				"name": "Groceries",
				"description": "Food and household items",
				"is_income": false,
				"exclude_from_budget": false,
				"exclude_from_totals": false,
				"updated_at": "2023-01-01T00:00:00Z",
				"created_at": "2023-01-01T00:00:00Z",
				"is_group": false,
				"group_id": null,
				"archived": false,
				"archived_at": null,
				"order": null,
				"collapsed": false
			}`,
			// The spec answers this GET with a 201.
			statusCode: http.StatusCreated,
			want: &Category{
				ID:                1,
				Name:              "Groceries",
				Description:       "Food and household items",
				IsIncome:          false,
				ExcludeFromBudget: false,
				ExcludeFromTotals: false,
				UpdatedAt:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				CreatedAt:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				IsGroup:           false,
			},
		},
		{
			name:        "invalid response",
			id:          1,
			response:    `{"invalid": "json"`,
			statusCode:  http.StatusOK,
			wantErr:     true,
			errContains: "decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/categories/1", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).GetCategory(context.Background(), tt.id)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
