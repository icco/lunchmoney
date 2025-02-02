package lunchmoney

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionFilters_ToMap(t *testing.T) {
	tagID := int64(1)
	recurringID := int64(2)
	plaidAccountID := int64(3)
	categoryID := int64(4)
	assetID := int64(5)
	offset := int64(10)
	limit := int64(20)
	startDate := "2023-01-01"
	endDate := "2023-12-31"
	debitAsNegative := true

	tests := []struct {
		name     string
		filters  TransactionFilters
		expected map[string]string
	}{
		{
			name: "all fields set",
			filters: TransactionFilters{
				TagID:           &tagID,
				RecurringID:     &recurringID,
				PlaidAccountID:  &plaidAccountID,
				CategoryID:      &categoryID,
				AssetID:         &assetID,
				Offset:          &offset,
				Limit:           &limit,
				StartDate:       &startDate,
				EndDate:         &endDate,
				DebitAsNegative: &debitAsNegative,
			},
			expected: map[string]string{
				"tag_id":            "1",
				"recurring_id":      "2",
				"plaid_account_id":  "3",
				"category_id":       "4",
				"asset_id":          "5",
				"offset":            "10",
				"limit":             "20",
				"start_date":        "2023-01-01",
				"end_date":          "2023-12-31",
				"debit_as_negative": "true",
			},
		},
		{
			name:     "no fields set",
			filters:  TransactionFilters{},
			expected: map[string]string{},
		},
		{
			name: "some fields set",
			filters: TransactionFilters{
				TagID:   &tagID,
				Limit:   &limit,
				EndDate: &endDate,
			},
			expected: map[string]string{
				"tag_id":   "1",
				"limit":    "20",
				"end_date": "2023-12-31",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filters.ToMap()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
