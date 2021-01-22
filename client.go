package lunchmoney

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Client holds our base configuration for our LunchMoney client.
type Client struct {
	Http *http.Client
	Base *url.URL
}

// NewClient creates a new client with the specified API keu
func NewClient(apikey string) (*Client, error) {
	base, err := url.Parse("https://dev.lunchmoney.app/v1/")
	if err != nil {
		return nil, fmt.Errorf("invalid base URI: %w", err)
	}

	return &Client{
		hc: &http.Client{
			Transport: &addHeaderTransport{T: http.DefaultTransport, Key: apikey},
		},
		Base: base,
	}, nil
}

// Get makes a request using the client to the path specified with the
// key/value pairs specified in options. It returns the body of the response or
// an error.
func (c *Client) Get(ctx context.Context, path string, options map[string]string) (io.Reader, error) {
	u, err := url.Parse(c.Base.String() + path)
	if err != nil {
		return nil, fmt.Errorf("bad path: %w", err)
	}

	query := u.Query()
	for k, v := range options {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
	}

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request (%+v) failed: %w", req, err)
	}

	return resp.Body, nil
}
