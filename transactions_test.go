package lunchmoney

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEndDate = "2023-12-31"

func TestTransactionFilters_ToMap(t *testing.T) {
	tagID := int64(1)
	recurringID := int64(2)
	plaidAccountID := int64(3)
	categoryID := int64(4)
	manualAccountID := int64(5)
	offset := int64(10)
	limit := int64(20)
	startDate := "2023-01-01"
	endDate := testEndDate
	status := StatusUnreviewed
	includePending := true

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
				ManualAccountID: &manualAccountID,
				Offset:          &offset,
				Limit:           &limit,
				StartDate:       &startDate,
				EndDate:         &endDate,
				Status:          &status,
				IncludePending:  &includePending,
			},
			expected: map[string]string{
				"tag_id":            "1",
				"recurring_id":      "2",
				"plaid_account_id":  "3",
				"category_id":       "4",
				"manual_account_id": "5",
				"offset":            "10",
				"limit":             "20",
				"start_date":        "2023-01-01",
				"end_date":          testEndDate,
				"status":            StatusUnreviewed,
				"include_pending":   "true",
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
				"end_date": testEndDate,
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

func TestGetTransactions(t *testing.T) {
	response := `{
		"transactions": [
			{
				"id": 2112150655,
				"date": "2024-07-28",
				"amount": "1250.8400",
				"currency": "usd",
				"to_base": 1250.84,
				"recurring_id": 994069,
				"payee": "Paycheck",
				"original_name": "DIRECT DEPOSIT PAYROLL",
				"category_id": null,
				"notes": null,
				"status": "reviewed",
				"is_pending": false,
				"created_at": "2024-07-28T17:00:06.192Z",
				"updated_at": "2024-07-28T17:00:06.733Z",
				"is_split_parent": false,
				"split_parent_id": null,
				"is_group_parent": false,
				"group_parent_id": null,
				"manual_account_id": null,
				"plaid_account_id": 119806,
				"tag_ids": [94317],
				"source": "plaid",
				"external_id": null
			}
		],
		"has_more": true
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).GetTransactions(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, got.Transactions, 1)
	assert.True(t, got.HasMore)

	tx := got.Transactions[0]
	assert.Nil(t, tx.CategoryID)
	assert.Nil(t, tx.ManualAccountID)
	assert.Equal(t, []int64{94317}, tx.TagIDs)
	assert.Equal(t, StatusReviewed, tx.Status)

	amount, err := tx.ParsedAmount()
	require.NoError(t, err)
	assert.Equal(t, int64(125084), amount.Amount())
}

func TestInsertTransactions(t *testing.T) {
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotBody))

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(`{
			"transactions": [{"id": 1, "amount": "10.0000", "currency": "usd"}],
			"skipped_duplicates": [
				{"reason": "duplicate_external_id", "request_transactions_index": 1, "existing_transaction_id": 99}
			]
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).InsertTransactions(context.Background(), InsertTransactionsRequest{
		Transactions: []InsertTransaction{
			{Date: "2024-07-28", Amount: "10.00", Status: StatusReviewed, TagIDs: []int64{5}},
		},
	})
	require.NoError(t, err)

	// Tags go up as IDs now; v2 will not create them inline.
	txs, ok := gotBody["transactions"].([]any)
	require.True(t, ok)
	require.Len(t, txs, 1)
	assert.Equal(t, []any{float64(5)}, txs[0].(map[string]any)["tag_ids"])

	require.Len(t, got.Transactions, 1)
	require.Len(t, got.SkippedDuplicates, 1)
	assert.Equal(t, "duplicate_external_id", got.SkippedDuplicates[0].Reason)
}

func TestInsertTransactionsRejectsOldStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("request should not have been made")
	}))
	defer server.Close()

	_, err := testClient(t, server).InsertTransactions(context.Background(), InsertTransactionsRequest{
		Transactions: []InsertTransaction{{Date: "2024-07-28", Amount: "10.00", Status: "cleared"}},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Status")
}

func TestUpdateTransaction(t *testing.T) {
	payee := "Rent"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions/42", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		// v2 takes the fields directly rather than under a "transaction" key.
		assert.JSONEq(t, `{"payee": "Rent"}`, string(body))

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(`{"id": 42, "payee": "Rent"}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).UpdateTransaction(context.Background(), 42, &UpdateTransaction{Payee: &payee})
	require.NoError(t, err)
	assert.Equal(t, int64(42), got.ID)
	assert.Equal(t, "Rent", got.Payee)
}
