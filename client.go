package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	// BaseAPIURL is the base url we use for all API requests.
	BaseAPIURL = "https://dev.lunchmoney.app/"
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
	HTTP *http.Client
	Base *url.URL
}

// NewClient creates a new client with the specified API key.
func NewClient(apikey string) (*Client, error) {
	base, err := url.Parse(BaseAPIURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URI: %w", err)
	}

	return &Client{
		HTTP: &http.Client{
			Transport: &addAuthHeaderTransport{T: http.DefaultTransport, Key: apikey},
		},
		Base: base,
	}, nil
}

// ErrorResponse is json if we get an error from the LM API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Get makes a request using the client to the path specified with the
// key/value pairs specified in options. It returns the body of the response or
// an error.
func (c *Client) Get(ctx context.Context, path string, options map[string]string) (io.Reader, error) {
	u, err := url.Parse(c.Base.String())
	if err != nil {
		return nil, fmt.Errorf("bad path: %w", err)
	}

	u.Path = path
	query := u.Query()
	for k, v := range options {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request (%+v) failed: %w", req, err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp *ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(errResp); err != nil {
			return nil, err
		}

		if errResp.Error != "" {
			return nil, fmt.Errorf("%s: %q", resp.Status, errResp.Error)
		}
	}

	return resp.Body, nil
}
