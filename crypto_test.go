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

func TestGetCryptocurrencies(t *testing.T) {
	response := `{"cryptocurrencies":[{"id":1,"coingecko_id":"bitcoin","symbol":"btc","full_name":"Bitcoin"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/cryptocurrencies", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).GetCryptocurrencies(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(1), got[0].ID)
	assert.Equal(t, "btc", got[0].Symbol)
}

func TestCreateCryptoManual(t *testing.T) {
	response := `{"id":22045,"name":"Coinbase ETH Holdings","display_name":"Trading ETH","institution_name":"Coinbase","balance":"12.004500000000000000","symbol":"eth","coingecko_id":"ethereum","to_base":28998.44,"balance_as_of":"2026-03-01T09:20:41.000Z","exchange_rate_as_of":"2026-03-01T09:15:00.000Z","created_by_name":"User 1","created_at":"2026-03-01T09:20:41.000Z","updated_at":"2026-03-01T09:20:41.000Z"}`

	var gotBody map[string]any

	displayName := "Trading ETH"
	institution := "Coinbase"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/crypto/manual", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotBody))

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).CreateCryptoManual(context.Background(), &CreateCryptoManual{
		Name:            "Coinbase ETH Holdings",
		DisplayName:     &displayName,
		InstitutionName: &institution,
		Balance:         "12.004500000000000000",
		Symbol:          "eth",
	})
	require.NoError(t, err)

	assert.Equal(t, "Coinbase ETH Holdings", gotBody["name"])
	assert.Equal(t, "12.004500000000000000", gotBody["balance"])
	assert.Equal(t, "eth", gotBody["symbol"])
	assert.NotContains(t, gotBody, "coingecko_id")

	require.NotNil(t, got)
	assert.Equal(t, int64(22045), got.ID)
	assert.Equal(t, "eth", got.Symbol)
}

func TestDeleteCryptoManual(t *testing.T) {
	yes := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/crypto/manual/22045", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("keep_history"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	err := testClient(t, server).DeleteCryptoManual(context.Background(), 22045, &DeleteCryptoManualOptions{KeepHistory: &yes})
	require.NoError(t, err)
}

func TestRefreshCryptoSynced(t *testing.T) {
	response := `{"id":33004,"provider":"coinbase","status":"active","created_by_name":"User 1","created_at":"2025-10-02T11:02:09.000Z","updated_at":"2026-02-25T14:25:01.000Z","display_name":"Coinbase Main","balances":[{"name":"ETH","display_name":null,"balance":"12.004500000000000000","symbol":"eth","coingecko_id":"ethereum","to_base":28998.44,"balance_as_of":"2026-02-25T14:25:00.000Z","exchange_rate_as_of":"2026-02-25T14:20:00.000Z","updated_at":"2026-02-25T14:25:01.000Z"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/crypto/synced/33004/refresh", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	got, err := testClient(t, server).RefreshCryptoSynced(context.Background(), 33004)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(33004), got.ID)
	require.Len(t, got.Balances, 1)
	assert.Equal(t, "eth", got.Balances[0].Symbol)
}
