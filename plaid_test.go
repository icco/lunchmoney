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

func TestPlaidFetchFilters_ToMap(t *testing.T) {
	startDate := "2023-01-01"
	endDate := "2023-12-31"
	id := int64(119807)

	tests := []struct {
		name     string
		filters  PlaidFetchFilters
		expected map[string]string
	}{
		{
			name: "all fields set",
			filters: PlaidFetchFilters{
				StartDate: &startDate,
				EndDate:   &endDate,
				ID:        &id,
			},
			expected: map[string]string{
				"start_date": "2023-01-01",
				"end_date":   "2023-12-31",
				"id":         "119807",
			},
		},
		{
			name:     "no fields set",
			filters:  PlaidFetchFilters{},
			expected: map[string]string{},
		},
		{
			name:     "id only",
			filters:  PlaidFetchFilters{ID: &id},
			expected: map[string]string{"id": "119807"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filters.ToMap()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestTriggerPlaidFetch(t *testing.T) {
	startDate := "2023-01-01"
	endDate := "2023-12-31"
	id := int64(119807)

	// The 425 body is the example the spec documents for the response.
	tooEarlyBody := `{"message": "Too Early", "errors": [{"errMsg": "Please wait at least 60 seconds between fetch requests."}]}`

	tests := []struct {
		name        string
		filters     *PlaidFetchFilters
		statusCode  int
		response    string
		wantQuery   url.Values
		wantIDs     []int64
		wantErr     bool
		wantTooEarl bool
	}{
		{
			name:       "accepted",
			statusCode: http.StatusAccepted,
			response:   `{"plaid_accounts": [{"id": 119807, "name": "Checking"}]}`,
			wantQuery:  url.Values{},
			wantIDs:    []int64{119807},
		},
		{
			name:       "accepted with no body",
			statusCode: http.StatusAccepted,
			response:   ``,
			wantQuery:  url.Values{},
			wantIDs:    nil,
		},
		{
			name: "filters become query parameters",
			filters: &PlaidFetchFilters{
				StartDate: &startDate,
				EndDate:   &endDate,
				ID:        &id,
			},
			statusCode: http.StatusAccepted,
			response:   `{"plaid_accounts": []}`,
			wantQuery: url.Values{
				"start_date": {"2023-01-01"},
				"end_date":   {"2023-12-31"},
				"id":         {"119807"},
			},
			wantIDs: []int64{},
		},
		{
			name:        "too early",
			statusCode:  http.StatusTooEarly,
			response:    tooEarlyBody,
			wantQuery:   url.Values{},
			wantErr:     true,
			wantTooEarl: true,
		},
		{
			name:       "other errors are not too early",
			statusCode: http.StatusBadRequest,
			response:   `{"message": "Invalid Request Parameters", "errors": [{"errMsg": "Both 'start_date' and 'end_date' must be specified."}]}`,
			wantQuery:  url.Values{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			gotQuery := url.Values{}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.Query()

				w.WriteHeader(tt.statusCode)
				if tt.response != "" {
					_, err := w.Write([]byte(tt.response))
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			accounts, err := testClient(t, server).TriggerPlaidFetch(context.Background(), tt.filters)

			assert.Equal(t, http.MethodPost, gotMethod)
			assert.Equal(t, "/plaid_accounts/fetch", gotPath)
			assert.Equal(t, tt.wantQuery, gotQuery)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantTooEarl, IsTooEarly(err))

				return
			}

			require.NoError(t, err)
			assert.False(t, IsTooEarly(err))

			ids := []int64{}
			for _, a := range accounts {
				ids = append(ids, a.ID)
			}
			if tt.wantIDs == nil {
				assert.Empty(t, accounts)
			} else {
				assert.Equal(t, tt.wantIDs, ids)
			}
		})
	}
}

func TestTriggerPlaidFetchValidatesDatePairing(t *testing.T) {
	// The API requires start_date and end_date together, so a lone date has to
	// fail before we ever send the request.
	date := "2023-01-01"
	badDate := "01/01/2023"

	tests := []struct {
		name    string
		filters *PlaidFetchFilters
		wantErr bool
	}{
		{name: "start date only", filters: &PlaidFetchFilters{StartDate: &date}, wantErr: true},
		{name: "end date only", filters: &PlaidFetchFilters{EndDate: &date}, wantErr: true},
		{name: "both dates", filters: &PlaidFetchFilters{StartDate: &date, EndDate: &date}},
		{name: "neither date", filters: &PlaidFetchFilters{}},
		{name: "bad date format", filters: &PlaidFetchFilters{StartDate: &badDate, EndDate: &date}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusAccepted)
			}))
			defer server.Close()

			_, err := testClient(t, server).TriggerPlaidFetch(context.Background(), tt.filters)
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, called)

				return
			}

			require.NoError(t, err)
			assert.True(t, called)
		})
	}
}
