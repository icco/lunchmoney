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
	tests := []struct {
		name        string
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
						"group_id": 0
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
					GroupID:           0,
				},
			},
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
				assert.Equal(t, "/v1/categories", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient("test-token")
			client.baseURL = server.URL

			got, err := client.GetCategories(context.Background())
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
				"group_id": 0
			}`,
			statusCode: http.StatusOK,
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
				GroupID:           0,
			},
		},
		{
			name:        "invalid response",
			id:          1,
			response:    `{"invalid": "json"`,
			statusCode:  http.StatusOK,
			wantErr:     true,
			errContains: "error getting category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/categories/1", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := NewClient("test-token")
			client.baseURL = server.URL

			got, err := client.GetCategory(context.Background(), tt.id)
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
