package lunchmoney

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// invalidPeriodBody is the error the API answers with when a start date is not
// a period start for the account.
const invalidPeriodBody = `{
	"message": "Invalid Request",
	"errMsg": "The requested start date is not a valid budget period start for this account.",
	"requested_start_date": "2025-01-15",
	"previous_valid_start_date": "2025-01-01",
	"next_valid_start_date": "2025-02-01"
}`

func TestUpsertBudget(t *testing.T) {
	notes := "Monthly groceries"
	cleared := ""
	tooLong := strings.Repeat("x", 351)

	tests := []struct {
		name        string
		budget      *UpsertBudget
		wantBody    map[string]any
		statusCode  int
		response    string
		errContains string
		wantPeriod  bool
		want        *Budget
	}{
		{
			name:   "new budget",
			budget: &UpsertBudget{StartDate: "2025-01-01", CategoryID: 315177, Amount: "500", Currency: "usd", Notes: &notes},
			wantBody: map[string]any{
				"start_date":  "2025-01-01",
				"category_id": float64(315177),
				"amount":      "500",
				"currency":    "usd",
				"notes":       "Monthly groceries",
			},
			response: `{"category_id": 315177, "start_date": "2025-01-01", "amount": "500.0000", "currency": "usd", "to_base": 500.0, "notes": "Monthly groceries"}`,
			want:     &Budget{CategoryID: 315177, StartDate: "2025-01-01", Amount: "500.0000", Currency: "usd", ToBase: 500, Notes: "Monthly groceries"},
		},
		{
			name:     "empty notes clear the stored ones",
			budget:   &UpsertBudget{StartDate: "2025-02-01", CategoryID: 315177, Amount: "42.5000", Notes: &cleared},
			wantBody: map[string]any{"start_date": "2025-02-01", "category_id": float64(315177), "amount": "42.5000", "notes": ""},
			response: `{"category_id": 315177, "start_date": "2025-02-01", "amount": "42.5000", "currency": "usd", "to_base": 42.5, "notes": ""}`,
			want:     &Budget{CategoryID: 315177, StartDate: "2025-02-01", Amount: "42.5000", Currency: "usd", ToBase: 42.5},
		},
		{
			name:        "missing start date is rejected before the request",
			budget:      &UpsertBudget{CategoryID: 315177, Amount: "500"},
			errContains: "StartDate",
		},
		{
			name:        "missing category is rejected before the request",
			budget:      &UpsertBudget{StartDate: "2025-01-01", Amount: "500"},
			errContains: "CategoryID",
		},
		{
			name:        "missing amount is rejected before the request",
			budget:      &UpsertBudget{StartDate: "2025-01-01", CategoryID: 315177},
			errContains: "Amount",
		},
		{
			name:        "a start date that is not a date is rejected before the request",
			budget:      &UpsertBudget{StartDate: "January 2025", CategoryID: 315177, Amount: "500"},
			errContains: "StartDate",
		},
		{
			name:        "too long a note is rejected before the request",
			budget:      &UpsertBudget{StartDate: "2025-01-01", CategoryID: 315177, Amount: "500", Notes: &tooLong},
			errContains: "Notes",
		},
		{
			name:        "start date is not a period start",
			budget:      &UpsertBudget{StartDate: "2025-01-15", CategoryID: 315177, Amount: "500"},
			wantBody:    map[string]any{"start_date": "2025-01-15", "category_id": float64(315177), "amount": "500"},
			statusCode:  http.StatusBadRequest,
			response:    invalidPeriodBody,
			errContains: "not a valid budget period start",
			wantPeriod:  true,
		},
		{
			name:        "unknown category",
			budget:      &UpsertBudget{StartDate: "2025-01-01", CategoryID: 1, Amount: "500"},
			wantBody:    map[string]any{"start_date": "2025-01-01", "category_id": float64(1), "amount": "500"},
			statusCode:  http.StatusBadRequest,
			response:    `{"message": "Invalid Request Body", "errors": [{"errMsg": "Category ID does not exist"}]}`,
			errContains: "Category ID does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/budgets", r.URL.Path)
				assert.Equal(t, http.MethodPut, r.Method)

				got := map[string]any{}
				require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
				assert.Equal(t, tt.wantBody, got)

				if tt.statusCode != 0 {
					w.WriteHeader(tt.statusCode)
				}

				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).UpsertBudget(context.Background(), tt.budget)
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assertInvalidPeriod(t, err, tt.wantPeriod)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDeleteBudget(t *testing.T) {
	tests := []struct {
		name        string
		categoryID  int64
		startDate   string
		statusCode  int
		response    string
		errContains string
		wantPeriod  bool
	}{
		{
			name:       "deleted",
			categoryID: 315177,
			startDate:  "2025-01-01",
			statusCode: http.StatusNoContent,
		},
		{
			name:        "a start date that is not a date is rejected before the request",
			categoryID:  315177,
			startDate:   "01/01/2025",
			errContains: "datetime",
		},
		{
			name:        "a missing start date is rejected before the request",
			categoryID:  315177,
			errContains: "required",
		},
		{
			name:        "start date is not a period start",
			categoryID:  315177,
			startDate:   "2025-01-15",
			statusCode:  http.StatusBadRequest,
			response:    invalidPeriodBody,
			errContains: "not a valid budget period start",
			wantPeriod:  true,
		},
		{
			name:        "unknown category",
			categoryID:  1,
			startDate:   "2025-01-01",
			statusCode:  http.StatusBadRequest,
			response:    `{"message": "Invalid Request Body", "errors": [{"errMsg": "Category ID does not exist"}]}`,
			errContains: "Category ID does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/budgets", r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)

				// The budget to delete is identified by query parameters,
				// which url.Values sorts by name.
				assert.Equal(t, "category_id="+strconv.FormatInt(tt.categoryID, 10)+"&start_date="+tt.startDate, r.URL.RawQuery)

				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			err := testClient(t, server).DeleteBudget(context.Background(), tt.categoryID, tt.startDate)
			if tt.errContains == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
			assertInvalidPeriod(t, err, tt.wantPeriod)
		})
	}
}

// assertInvalidPeriod checks whether err carries the period details, and that
// anything that does still unwraps to the API error.
func assertInvalidPeriod(t *testing.T, err error, want bool) {
	t.Helper()

	var periodErr *BudgetInvalidPeriodError
	if !want {
		assert.False(t, errors.As(err, &periodErr))
		return
	}

	require.True(t, errors.As(err, &periodErr))
	assert.Equal(t, "2025-01-15", periodErr.RequestedStartDate)
	assert.Equal(t, "2025-01-01", periodErr.PreviousValidStartDate)
	assert.Equal(t, "2025-02-01", periodErr.NextValidStartDate)

	var apiErr *ErrorResponse
	require.True(t, errors.As(err, &apiErr))
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
}
