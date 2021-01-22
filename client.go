package lunchmoney

import (
	"context"
	"fmt"
	"net/http"
)

type addAuthHeaderTransport struct {
	T   http.RoundTripper
	Key string
}

func (adt *addAuthHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if adt.Key == "" {
		return nil, fmt.Errorf("no key provided")
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", adt.Key))
	req.Header.Add("User-Agent", "github.com/icco/lunchmoney/0.0.0")

	return adt.T.RoundTrip(req)
}

type Client struct {
	hc *http.Client
}

func NewClient(ctx context.Context, apikey string) (*Client, error) {
	return &Client{
		hc: &http.Client{
			Transport: &addHeaderTransport{T: http.DefaultTransport, Key: apikey},
		},
	}, nil
}
