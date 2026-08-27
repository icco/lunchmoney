package lunchmoney

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// splitParentResponse is a split parent with its two new children.
const splitParentResponse = `{
	"id": 42,
	"amount": "88.4500",
	"currency": "usd",
	"payee": "Food Town",
	"is_split_parent": true,
	"children": [
		{"id": 43, "amount": "44.2300", "currency": "usd", "payee": "Food Town - Lenny"},
		{"id": 44, "amount": "44.2200", "currency": "usd", "payee": "Food Town - Penny"}
	]
}`

// groupParentResponse is a group parent with its two members.
const groupParentResponse = `{
	"id": 99,
	"amount": "375.0000",
	"currency": "usd",
	"payee": "Home Entertainment Transactions",
	"is_group_parent": true,
	"children": [
		{"id": 1, "amount": "75.0000", "currency": "usd", "group_parent_id": 99},
		{"id": 2, "amount": "300.0000", "currency": "usd", "group_parent_id": 99}
	]
}`

func TestSplitTransaction(t *testing.T) {
	categoryID := int64(315162)

	tests := []struct {
		name     string
		req      SplitTransactionRequest
		wantBody string
	}{
		{
			name: "two children",
			req: SplitTransactionRequest{
				ChildTransactions: []SplitTransactionChild{
					{Amount: "44.23", Payee: "Food Town - Lenny"},
					{Amount: "44.22", Payee: "Food Town - Penny"},
				},
			},
			wantBody: `{"child_transactions": [
				{"amount": "44.23", "payee": "Food Town - Lenny"},
				{"amount": "44.22", "payee": "Food Town - Penny"}
			]}`,
		},
		{
			name: "optional fields",
			req: SplitTransactionRequest{
				ChildTransactions: []SplitTransactionChild{
					{Amount: "44.23", Date: "2024-10-19", CategoryID: &categoryID, TagIDs: []int64{7}, Notes: "half"},
					{Amount: "44.22"},
				},
			},
			wantBody: `{"child_transactions": [
				{"amount": "44.23", "date": "2024-10-19", "category_id": 315162, "tag_ids": [7], "notes": "half"},
				{"amount": "44.22"}
			]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/transactions/split/42", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.wantBody, string(body))

				w.WriteHeader(http.StatusCreated)
				_, err = w.Write([]byte(splitParentResponse))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).SplitTransaction(context.Background(), 42, tt.req)
			require.NoError(t, err)
			assert.Equal(t, int64(42), got.ID)
			assert.True(t, got.IsSplitParent)
			require.Len(t, got.Children, 2)
			assert.Equal(t, "Food Town - Lenny", got.Children[0].Payee)
		})
	}
}

func TestSplitTransactionValidation(t *testing.T) {
	tests := []struct {
		name        string
		req         SplitTransactionRequest
		errContains string
	}{
		{
			name:        "no children",
			req:         SplitTransactionRequest{},
			errContains: "ChildTransactions",
		},
		{
			name: "one child",
			req: SplitTransactionRequest{
				ChildTransactions: []SplitTransactionChild{{Amount: "88.45"}},
			},
			errContains: "ChildTransactions",
		},
		{
			name: "child without an amount",
			req: SplitTransactionRequest{
				ChildTransactions: []SplitTransactionChild{{Amount: "44.23"}, {Payee: "Food Town - Penny"}},
			},
			errContains: "Amount",
		},
		{
			name: "child with a bad date",
			req: SplitTransactionRequest{
				ChildTransactions: []SplitTransactionChild{
					{Amount: "44.23", Date: "10/19/2024"},
					{Amount: "44.22"},
				},
			},
			errContains: "Date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				t.Error("request should not have been made")
			}))
			defer server.Close()

			_, err := testClient(t, server).SplitTransaction(context.Background(), 42, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestUnsplitTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions/split/42", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	require.NoError(t, testClient(t, server).UnsplitTransaction(context.Background(), 42))
}

func TestGroupTransactions(t *testing.T) {
	categoryID := int64(315628)

	tests := []struct {
		name     string
		req      GroupTransactionsRequest
		wantBody string
	}{
		{
			name: "two ids",
			req: GroupTransactionsRequest{
				IDs:   []int64{1, 2},
				Date:  "2024-12-10",
				Payee: "Home Entertainment Transactions",
			},
			// v2 takes the IDs under "ids"; v1 used "transactions".
			wantBody: `{"ids": [1, 2], "date": "2024-12-10", "payee": "Home Entertainment Transactions"}`,
		},
		{
			name: "empty payee is still sent",
			req: GroupTransactionsRequest{
				IDs:  []int64{1, 2},
				Date: "2024-12-10",
			},
			// The API requires the key and allows an empty value, so it must
			// not be dropped.
			wantBody: `{"ids": [1, 2], "date": "2024-12-10", "payee": ""}`,
		},
		{
			name: "optional fields",
			req: GroupTransactionsRequest{
				IDs:        []int64{1, 2},
				Date:       "2024-12-10",
				Payee:      "Home Entertainment Transactions",
				CategoryID: &categoryID,
				Notes:      "December",
				Status:     StatusUnreviewed,
				TagIDs:     []int64{94318},
			},
			wantBody: `{
				"ids": [1, 2],
				"date": "2024-12-10",
				"payee": "Home Entertainment Transactions",
				"category_id": 315628,
				"notes": "December",
				"status": "unreviewed",
				"tag_ids": [94318]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/transactions/group", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.wantBody, string(body))

				w.WriteHeader(http.StatusCreated)
				_, err = w.Write([]byte(groupParentResponse))
				require.NoError(t, err)
			}))
			defer server.Close()

			got, err := testClient(t, server).GroupTransactions(context.Background(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, int64(99), got.ID)
			assert.True(t, got.IsGroupParent)
			require.Len(t, got.Children, 2)
			require.NotNil(t, got.Children[0].GroupParentID)
			assert.Equal(t, int64(99), *got.Children[0].GroupParentID)
		})
	}
}

func TestGroupTransactionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		req         GroupTransactionsRequest
		errContains string
	}{
		{
			name:        "one id",
			req:         GroupTransactionsRequest{IDs: []int64{1}, Date: "2024-12-10"},
			errContains: "IDs",
		},
		{
			name:        "no date",
			req:         GroupTransactionsRequest{IDs: []int64{1, 2}},
			errContains: "Date",
		},
		{
			name:        "bad status",
			req:         GroupTransactionsRequest{IDs: []int64{1, 2}, Date: "2024-12-10", Status: "cleared"},
			errContains: "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				t.Error("request should not have been made")
			}))
			defer server.Close()

			_, err := testClient(t, server).GroupTransactions(context.Background(), tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestUngroupTransactions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transactions/group/99", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	require.NoError(t, testClient(t, server).UngroupTransactions(context.Background(), 99))
}

func TestUngroupTransactionsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, err := w.Write([]byte(`{"message": "Not Found", "errors": [{"errMsg": "There is no transaction with the id: 543210"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	err := testClient(t, server).UngroupTransactions(context.Background(), 543210)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ungroup transactions 543210")
	assert.Contains(t, err.Error(), "There is no transaction with the id: 543210")
}
