package lunchmoney

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalanceHistoryFiltersToMap(t *testing.T) {
	start := "2026-01"
	end := "2026-03"
	bad := "2026-01-01"

	tests := []struct {
		name        string
		filters     *BalanceHistoryFilters
		want        map[string]string
		errContains string
	}{
		{name: "nil filters", filters: nil, want: map[string]string{}},
		{name: "empty filters", filters: &BalanceHistoryFilters{}, want: map[string]string{}},
		{name: "both months set", filters: &BalanceHistoryFilters{StartMonth: &start, EndMonth: &end}, want: map[string]string{"start_month": "2026-01", "end_month": "2026-03"}},
		{name: "missing end month", filters: &BalanceHistoryFilters{StartMonth: &start}, errContains: "both be provided"},
		{name: "invalid start month", filters: &BalanceHistoryFilters{StartMonth: &bad, EndMonth: &end}, errContains: "YYYY-MM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filters.ToMap()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBalanceHistory(t *testing.T) {
	start := "2026-01"
	end := "2026-03"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/balance_history", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "2026-01", r.URL.Query().Get("start_month"))
		assert.Equal(t, "2026-03", r.URL.Query().Get("end_month"))

		_, err := w.Write([]byte(`{"balance_history":[{"source":{"type":"manual","manual_account_id":119807},"balances":[{"type":"historical","id":201,"month":"2026-01","balance":"41000.0000","currency":"usd","to_base":41000,"crypto_balance":null}]}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).GetBalanceHistory(context.Background(), &BalanceHistoryFilters{StartMonth: &start, EndMonth: &end})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "manual", got[0].Source.Type)
	require.Len(t, got[0].Balances, 1)
	assert.Equal(t, "2026-01", got[0].Balances[0].Month)
}

func TestGetBalanceHistoryForAccountValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request should not have been made")
	}))
	defer server.Close()

	_, err := testClient(t, server).GetBalanceHistoryForAccount(context.Background(), BalanceHistoryAccountType("bad"), 1, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "oneof")
}

func TestUpsertBalanceHistoryForCryptoSynced(t *testing.T) {
	response := `{"source":{"type":"crypto_synced","crypto_synced_id":33004,"symbol":"btc"},"balances":[{"type":"historical","id":604,"month":"2026-03","balance":"6400.0000","currency":"usd","to_base":6400,"crypto_balance":"0.100020003000400050"}]}`

	var gotBody map[string]any
	cryptoBalance := "0.100020003000400050"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/balance_history/crypto_synced/33004/btc", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).UpsertBalanceHistoryForCryptoSynced(context.Background(), 33004, "btc", &UpsertBalanceHistory{
		Balances: []*UpsertBalanceHistoryItem{{
			Month:         "2026-03",
			Balance:       "6400.0000",
			CryptoBalance: &cryptoBalance,
		}},
	})
	require.NoError(t, err)

	balances, ok := gotBody["balances"].([]any)
	require.True(t, ok)
	require.Len(t, balances, 1)
	item, ok := balances[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "2026-03", item["month"])
	assert.Equal(t, "6400.0000", item["balance"])

	require.NotNil(t, got)
	assert.Equal(t, "crypto_synced", got.Source.Type)
	require.Len(t, got.Balances, 1)
	assert.Equal(t, "2026-03", got.Balances[0].Month)
}

func TestDeleteBalanceHistoryForAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/balance_history/manual/119807", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := testClient(t, server).DeleteBalanceHistoryForAccount(context.Background(), BalanceHistoryAccountTypeManual, 119807)
	require.NoError(t, err)
}
