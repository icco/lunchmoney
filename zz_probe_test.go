package lunchmoney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbeInsertValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"transactions":[]}`))
	}))
	defer server.Close()
	c := testClient(t, server)

	// invalid date + missing amount in a nested element
	_, err := c.InsertTransactions(context.Background(), InsertTransactionsRequest{
		Transactions: []InsertTransaction{{Date: "not-a-date"}},
	})
	t.Logf("nested invalid -> err=%v", err)

	// empty slice
	_, err = c.InsertTransactions(context.Background(), InsertTransactionsRequest{})
	t.Logf("empty slice -> err=%v", err)
}

func TestProbeFilters(t *testing.T) {
	bad := "2024-13-45"
	f := &TransactionFilters{StartDate: &bad}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("REQ URL: %s", r.URL.String())
		_, _ = w.Write([]byte(`{"transactions":[]}`))
	}))
	defer server.Close()
	c := testClient(t, server)
	_, err := c.GetTransactions(context.Background(), f)
	t.Logf("bad start date -> err=%v", err)

	st := "bogus"
	_, err = c.GetTransactions(context.Background(), &TransactionFilters{Status: &st})
	t.Logf("bad status -> err=%v", err)
}

func TestProbeNilUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()
	c := testClient(t, server)
	_, err := c.UpdateTransaction(context.Background(), 1, nil)
	t.Logf("nil update -> err=%v", err)
}

func TestProbeJoinPath(t *testing.T) {
	var got string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.String()
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	c := testClient(t, server)
	_, _ = c.Get(context.Background(), "/transactions", map[string]string{"start_date": "2024-01-01"})
	t.Logf("url=%s", got)
}
