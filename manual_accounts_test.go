package lunchmoney

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateManualAccount(t *testing.T) {
	response := `{
		"id": 119999,
		"name": "API created loan",
		"institution_name": "Bank of America",
		"display_name": "Car loan",
		"type": "vehicle",
		"subtype": "loan",
		"balance": "9999.99",
		"currency": "usd",
		"to_base": 9999.99,
		"balance_as_of": "2024-10-12",
		"status": "active",
		"closed_on": null,
		"external_id": null,
		"exclude_from_transactions": false,
		"created_by_name": "User 1",
		"created_at": "2024-10-06T18:57:12.029Z",
		"updated_at": "2024-10-06T18:57:12.029Z"
	}`

	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/manual_accounts", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotBody))

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(response))
		require.NoError(t, err)
	}))
	defer server.Close()

	institution := "Bank of America"
	displayName := "Car loan"
	subtype := "loan"
	balanceAsOf := "2024-10-12"

	got, err := testClient(t, server).CreateManualAccount(context.Background(), &CreateManualAccount{
		Name:            "API created loan",
		Type:            "vehicle",
		Balance:         "9999.99",
		InstitutionName: &institution,
		DisplayName:     &displayName,
		Subtype:         &subtype,
		BalanceAsOf:     &balanceAsOf,
	})
	require.NoError(t, err)

	// Unset optional fields stay out of the request body.
	assert.Equal(t, "API created loan", gotBody["name"])
	assert.Equal(t, "vehicle", gotBody["type"])
	assert.Equal(t, "9999.99", gotBody["balance"])
	assert.Equal(t, balanceAsOf, gotBody["balance_as_of"])
	assert.NotContains(t, gotBody, "status")
	assert.NotContains(t, gotBody, "external_id")

	assert.Equal(t, int64(119999), got.ID)
	assert.Equal(t, "Car loan", got.DisplayName)
	// balance_as_of arrives here as a plain date rather than a date-time.
	assert.Equal(t, 2024, got.BalanceAsOf.Year())

	amount, err := got.ParsedAmount()
	require.NoError(t, err)
	assert.Equal(t, int64(999999), amount.Amount())
}

func TestCreateManualAccountValidation(t *testing.T) {
	longID := strings.Repeat("a", 76)
	badStatus := "archived"

	tests := []struct {
		name        string
		account     *CreateManualAccount
		errContains string
	}{
		{
			name:        "missing name",
			account:     &CreateManualAccount{Type: "cash", Balance: "100"},
			errContains: "Name",
		},
		{
			name:        "missing balance",
			account:     &CreateManualAccount{Name: "Savings", Type: "cash"},
			errContains: "Balance",
		},
		{
			name:        "v1 account type",
			account:     &CreateManualAccount{Name: "Savings", Type: "depository", Balance: "100"},
			errContains: "Type",
		},
		{
			name:        "external id too long",
			account:     &CreateManualAccount{Name: "Savings", Type: "cash", Balance: "100", ExternalID: &longID},
			errContains: "ExternalID",
		},
		{
			name:        "unknown status",
			account:     &CreateManualAccount{Name: "Savings", Type: "cash", Balance: "100", Status: &badStatus},
			errContains: "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				t.Error("request should not have been made")
			}))
			defer server.Close()

			_, err := testClient(t, server).CreateManualAccount(context.Background(), tt.account)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestDeleteManualAccount(t *testing.T) {
	yes := true
	no := false

	tests := []struct {
		name  string
		opts  *DeleteManualAccountOptions
		query map[string]string
	}{
		{
			name:  "no options",
			opts:  nil,
			query: map[string]string{"delete_items": "", "delete_balance_history": ""},
		},
		{
			name:  "delete items only",
			opts:  &DeleteManualAccountOptions{DeleteItems: &yes},
			query: map[string]string{"delete_items": "true", "delete_balance_history": ""},
		},
		{
			name:  "both options",
			opts:  &DeleteManualAccountOptions{DeleteItems: &no, DeleteBalanceHistory: &yes},
			query: map[string]string{"delete_items": "false", "delete_balance_history": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/manual_accounts/119807", r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)
				for k, v := range tt.query {
					assert.Equal(t, v, r.URL.Query().Get(k))
				}

				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()

			require.NoError(t, testClient(t, server).DeleteManualAccount(context.Background(), 119807, tt.opts))
		})
	}
}

func TestDeleteManualAccountNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(`{"message": "Not Found", "errors": [{"errMsg": "There is no manual account with the id: 543210."}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	err := testClient(t, server).DeleteManualAccount(context.Background(), 543210, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete manual account 543210")
	assert.Contains(t, err.Error(), "There is no manual account with the id: 543210.")
}
