package lunchmoney

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecurringExpenseFilters_ToMap(t *testing.T) {
	tests := []struct {
		name     string
		filters  RecurringExpenseFilters
		expected map[string]string
	}{
		{
			name:     "zero value",
			filters:  RecurringExpenseFilters{},
			expected: map[string]string{"debit_as_negative": "false"},
		},
		{
			name:    "start date only",
			filters: RecurringExpenseFilters{StartDate: "2023-01-01"},
			expected: map[string]string{
				"start_date":        "2023-01-01",
				"debit_as_negative": "false",
			},
		},
		{
			name:    "all fields set",
			filters: RecurringExpenseFilters{StartDate: "2023-01-01", DebitAsNegative: true},
			expected: map[string]string{
				"start_date":        "2023-01-01",
				"debit_as_negative": "true",
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
