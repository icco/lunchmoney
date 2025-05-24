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
	ErrorString any   `json:"error,omitempty"`
	ErrorsArray []any `json:"errors,omitempty"`
}

func (e *ErrorResponse) Error() string {
	if e.ErrorString != nil {
		return fmt.Sprintf("%v", e.ErrorString)
	}

	msg := ""
	if len(e.ErrorsArray) > 0 {
		msg = fmt.Sprintf("%v", e.ErrorsArray)
	}

	return msg
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("error closing response body: %w: %w", cerr, err)
			} else {
				err = fmt.Errorf("error closing response body: %w", cerr)
			}
		}
	}()

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

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	return &buf, nil
}

// Put performs an HTTP PUT request to the specified API endpoint with the provided body.
// It returns the response body as an io.Reader or an error if the request fails.
func (c *Client) Put(ctx context.Context, path string, body any) (io.Reader, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// Post performs an HTTP POST request to the specified API endpoint with the provided body.
// It returns the response body as an io.Reader or an error if the request fails.
func (c *Client) Post(ctx context.Context, path string, body any) (io.Reader, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) do(ctx context.Context, method string, path string, body any) (io.Reader, error) {
	u, err := url.Parse(c.Base.String())
	if err != nil {
		return nil, fmt.Errorf("bad path: %w", err)
	}

	u.Path = path

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("could not marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request (%+v) failed: %w", req, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("error closing response body: %w: %w", cerr, err)
			} else {
				err = fmt.Errorf("error closing response body: %w", cerr)
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		err := c.tryToFindError(resp, &buf, true)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%s", resp.Status)
	}

	// Sometimes 200 still means that there is an error
	var finalReader bytes.Buffer
	err = c.tryToFindError(resp, &finalReader, false)
	if err != nil {
		return nil, err
	}

	return &finalReader, nil
}

func (*Client) tryToFindError(resp *http.Response, outBuf *bytes.Buffer, failOnDecodeErr bool) error {
	tee := io.TeeReader(resp.Body, outBuf)
	errResp := ErrorResponse{}
	if err := json.NewDecoder(tee).Decode(&errResp); err != nil {
		if failOnDecodeErr {
			return fmt.Errorf("could not decode error response %s: %w", outBuf.String(), err)
		}
		// some other message is involved here (eg array)
		return nil
	}

	if errResp.Error() != "" {
		return fmt.Errorf("%s: %s", resp.Status, errResp.Error())
	}
	return nil
}

// ParseCurrency converts a string amount and currency code into a money.Money struct.
// It parses the amount as a float, multiplies by 100 to convert to cents, and returns
// a Money object in the specified currency. Returns an error if the amount can't be parsed.
func ParseCurrency(amount, currency string) (*money.Money, error) {
	f, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, fmt.Errorf("%q is not valid float: %w", amount, err)
	}

	v := int64(100 * f)
	return money.New(v, currency), nil
}
