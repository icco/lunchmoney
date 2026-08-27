package lunchmoney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClient returns a client pointed at ts.
func testClient(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()

	client, err := NewClient("test-token")
	require.NoError(t, err)

	client.Base, err = url.Parse(ts.URL)
	require.NoError(t, err)

	return client
}

func TestClientKeepsBasePath(t *testing.T) {
	// Requests have to land under the base URL's /v2 prefix, not replace it.
	client, err := NewClient("test-token")
	require.NoError(t, err)

	got := client.Base.JoinPath("/transactions/42").String()
	assert.Equal(t, "https://api.lunchmoney.dev/v2/transactions/42", got)
}

func TestClientAcceptsCreated(t *testing.T) {
	// v2 answers writes with 201, and GET /categories/{id} with 201 too.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`{"id": 1}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	body, err := testClient(t, server).Post(context.Background(), "/transactions", map[string]string{})
	require.NoError(t, err)
	assert.NotNil(t, body)
}

func TestClientErrors(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		response    string
		errContains string
	}{
		{
			name:        "message only",
			statusCode:  http.StatusBadRequest,
			response:    `{"message": "end_date cannot be before start_date", "errors": []}`,
			errContains: "end_date cannot be before start_date",
		},
		{
			name:        "message and errors",
			statusCode:  http.StatusBadRequest,
			response:    `{"message": "Invalid request", "errors": [{"errMsg": "date is required", "field": "date"}]}`,
			errContains: "Invalid request: date is required",
		},
		{
			name:        "unauthorized with no body",
			statusCode:  http.StatusUnauthorized,
			response:    ``,
			errContains: "401 Unauthorized",
		},
		{
			name:        "non json body",
			statusCode:  http.StatusInternalServerError,
			response:    `<html>nope</html>`,
			errContains: "nope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(tt.response))
				require.NoError(t, err)
			}))
			defer server.Close()

			_, err := testClient(t, server).Get(context.Background(), "/me", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		currency string
		want     int64
		wantErr  bool
	}{
		// v2 sends four decimal places, so the extra precision has to round
		// rather than fall off the end of a float multiplication.
		{name: "four decimals", amount: "1250.8400", currency: "usd", want: 125084},
		{name: "rounds up", amount: "10.005", currency: "usd", want: 1001},
		{name: "rounds down", amount: "10.0049", currency: "usd", want: 1000},
		{name: "negative rounds away from zero", amount: "-0.005", currency: "usd", want: -1},
		{name: "no decimals", amount: "42", currency: "usd", want: 4200},
		{name: "one decimal", amount: "42.5", currency: "usd", want: 4250},
		{name: "leading dot", amount: ".75", currency: "usd", want: 75},
		{name: "zero decimal currency", amount: "1250.84", currency: "jpy", want: 1251},
		{name: "three decimal currency", amount: "1.2345", currency: "bhd", want: 1235},
		{name: "unknown currency defaults to two", amount: "1.234", currency: "zzz", want: 123},
		{name: "uppercase code", amount: "1.23", currency: "USD", want: 123},
		{name: "not a number", amount: "banana", currency: "usd", wantErr: true},
		{name: "empty", amount: "", currency: "usd", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCurrency(tt.amount, tt.currency)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Amount())
		})
	}
}
