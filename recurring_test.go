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

	monthly, err := got[0].MonthlyAmount()
	require.NoError(t, err)
	assert.Equal(t, 1250.84, monthly)
}

func TestRecurringCriteria_MonthlyFactorAndCadence(t *testing.T) {
	tests := []struct {
		name        string
		granularity string
		quantity    int64
		wantFactor  float64
		wantCadence string
		wantOk      bool
	}{
		{
			name:        "monthly",
			granularity: "month",
			quantity:    1,
			wantFactor:  1.0,
			wantCadence: "monthly",
			wantOk:      true,
		},
		{
			name:        "quarterly",
			granularity: "month",
			quantity:    3,
			wantFactor:  1.0 / 3.0,
			wantCadence: "every 3 months",
			wantOk:      true,
		},
		{
			name:        "weekly",
			granularity: "week",
			quantity:    1,
			wantFactor:  52.0 / 12.0,
			wantCadence: "weekly",
			wantOk:      true,
		},
		{
			name:        "biweekly",
			granularity: "week",
			quantity:    2,
			wantFactor:  26.0 / 12.0,
			wantCadence: "every 2 weeks",
			wantOk:      true,
		},
		{
			name:        "yearly",
			granularity: "year",
			quantity:    1,
			wantFactor:  1.0 / 12.0,
			wantCadence: "yearly",
			wantOk:      true,
		},
		{
			name:        "daily",
			granularity: "day",
			quantity:    1,
			wantFactor:  365.0 / 12.0,
			wantCadence: "daily",
			wantOk:      true,
		},
		{
			name:        "invalid quantity",
			granularity: "month",
			quantity:    0,
			wantFactor:  0,
			wantCadence: "every 0 months",
			wantOk:      false,
		},
		{
			name:        "unknown granularity",
			granularity: "unknown",
			quantity:    1,
			wantFactor:  0,
			wantCadence: "unknown",
			wantOk:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crit := RecurringCriteria{
				Granularity: tt.granularity,
				Quantity:    tt.quantity,
			}
			factor, ok := crit.MonthlyFactor()
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.InDelta(t, tt.wantFactor, factor, 0.0001)
			}
			assert.Equal(t, tt.wantCadence, crit.Cadence())
		})
	}
}

func TestRecurringItem_MonthlyAmount(t *testing.T) {
	item := &RecurringItem{
		TransactionCriteria: RecurringCriteria{
			Granularity: "week",
			Quantity:    2,
			Amount:      "100.0000",
		},
	}

	monthly, err := item.MonthlyAmount()
	require.NoError(t, err)
	// 100 * (26 / 12) = 216.67
	assert.Equal(t, 216.67, monthly)

	invalid := &RecurringItem{
		TransactionCriteria: RecurringCriteria{
			Granularity: "unknown",
			Quantity:    1,
			Amount:      "100.0000",
		},
	}
	_, err = invalid.MonthlyAmount()
	require.Error(t, err)
}
