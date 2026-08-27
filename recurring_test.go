package lunchmoney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurringItemFilters_ToMap(t *testing.T) {
	includeSuggested := true

	tests := []struct {
		name     string
		filters  RecurringItemFilters
		expected map[string]string
	}{
		{
			name:     "zero value",
			filters:  RecurringItemFilters{},
			expected: map[string]string{},
		},
		{
			name:    "date range",
			filters: RecurringItemFilters{StartDate: "2023-01-01", EndDate: testEndDate},
			expected: map[string]string{
				"start_date": "2023-01-01",
				"end_date":   testEndDate,
			},
		},
		{
			name: "all fields set",
			filters: RecurringItemFilters{
				StartDate:        "2023-01-01",
				EndDate:          testEndDate,
				IncludeSuggested: &includeSuggested,
			},
			expected: map[string]string{
				"start_date":        "2023-01-01",
				"end_date":          testEndDate,
				"include_suggested": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filters.ToMap()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetRecurringItemsRequiresBothDates(t *testing.T) {
	// v2 rejects a start date without an end date, so catch it before the call.
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("request should not have been made")
	}))
	defer server.Close()

	_, err := testClient(t, server).GetRecurringItems(context.Background(), &RecurringItemFilters{StartDate: "2023-01-01"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "EndDate")
}

func TestGetRecurringItems(t *testing.T) {
	response := `{
		"recurring_items": [
			{
				"id": 994069,
				"description": "Paycheck",
				"status": "reviewed",
				"transaction_criteria": {
					"start_date": null,
					"end_date": null,
					"granularity": "month",
					"quantity": 1,
					"anchor_date": "2024-07-28",
					"payee": "Paycheck",
					"amount": "1250.8400",
					"to_base": 1250.84,
					"currency": "usd",
					"plaid_account_id": 119806,
					"manual_account_id": null
				},
				"overrides": {"payee": "Paycheck"},
				"matches": null,
				"created_by": 1,
				"created_at": "2024-07-28T17:00:06.192Z",
				"updated_at": "2024-07-28T17:00:06.733Z",
				"source": "system"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/recurring_items", r.URL.Path)
		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).GetRecurringItems(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, int64(994069), got[0].ID)
	assert.Equal(t, "month", got[0].TransactionCriteria.Granularity)
	assert.Equal(t, "2024-07-28", got[0].TransactionCriteria.AnchorDate)
	assert.Nil(t, got[0].TransactionCriteria.ManualAccountID)
	assert.Nil(t, got[0].Matches)

	amount, err := got[0].ParsedAmount()
	require.NoError(t, err)
	assert.Equal(t, int64(125084), amount.Amount())
}
