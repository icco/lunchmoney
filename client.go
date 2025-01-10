package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Rhymond/go-money"
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
	req.Header.Add("User-Agent", "github.com/rshep3087/lunchmoney/0.0.0")

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
	ErrorString   string `json:"error,omitempty"`
	ErrorName     string `json:"name,omitempty"`
	MessageString string `json:"message,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if e.ErrorString != "" {
		return e.ErrorString
	}

	if e.MessageString != "" {
		return e.MessageString
	}

	if e.ErrorName != "" {
		return e.ErrorName
	}

	return ""
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
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)
		errResp := ErrorResponse{}
		if err := json.NewDecoder(tee).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("could not decode error response %s: %w", buf.String(), err)
		}

		// log.Printf("%s -> %+v", buf.String(), errResp)
		if errResp.Error() != "" {
			return nil, fmt.Errorf("%s: %s", resp.Status, errResp.Error())
		}

		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp.Body, nil
}

// ParseCurrency turns two strings into a money struct.
func ParseCurrency(amount, currency string) (*money.Money, error) {
	f, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("%q is not valid float: %w", amount, err)
	}

	v := int64(100 * f)
	return money.New(v, currency), nil
}
