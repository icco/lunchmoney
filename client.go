package lunchmoney

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
)

const (
	// BaseAPIURL is the base url we use for all API requests. It points at v2
	// of the Lunch Money API; v1 is not supported by this library.
	BaseAPIURL = "https://api.lunchmoney.dev/v2/"

	// userAgent identifies this library to the API.
	userAgent = "github.com/icco/lunchmoney"
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
	req.Header.Add("User-Agent", userAgent)

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

// ErrorResponse is the body the v2 API returns alongside a 4xx or 5xx status.
// Failing calls wrap one, so errors.As gets at the status code and the
// individual problems the API reported.
type ErrorResponse struct {
	Message string       `json:"message"`
	Errors  []ErrorEntry `json:"errors,omitempty"`

	// StatusCode is the HTTP status the body arrived with.
	StatusCode int `json:"-"`

	// RawBody is the undecoded error body, for the endpoints that answer a
	// failure with a shape of their own rather than message and errors.
	RawBody []byte `json:"-"`
}

// ErrorEntry is a single problem reported inside an ErrorResponse. The API is
// free to add fields per error type, so unrecognized keys are kept in Extra.
type ErrorEntry struct {
	Message string         `json:"errMsg"`
	Extra   map[string]any `json:"-"`
}

// UnmarshalJSON decodes an error entry, keeping any fields beyond errMsg in Extra.
func (e *ErrorEntry) UnmarshalJSON(b []byte) error {
	raw := map[string]any{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	if msg, ok := raw["errMsg"].(string); ok {
		e.Message = msg
	}
	delete(raw, "errMsg")

	if len(raw) > 0 {
		e.Extra = raw
	}

	return nil
}

func (e *ErrorResponse) Error() string {
	msgs := make([]string, 0, len(e.Errors)+1)
	if e.Message != "" {
		msgs = append(msgs, e.Message)
	}

	for _, entry := range e.Errors {
		if entry.Message != "" {
			msgs = append(msgs, entry.Message)
		}
	}

	return strings.Join(msgs, ": ")
}

// Get performs a GET request against the given path, with options as query parameters.
func (c *Client) Get(ctx context.Context, path string, options map[string]string) (io.Reader, error) {
	return c.do(ctx, http.MethodGet, path, options, nil)
}

// Put performs a PUT request against the given path with body as JSON.
func (c *Client) Put(ctx context.Context, path string, body any) (io.Reader, error) {
	return c.do(ctx, http.MethodPut, path, nil, body)
}

// Post performs a POST request against the given path with body as JSON.
func (c *Client) Post(ctx context.Context, path string, body any) (io.Reader, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

// Delete performs a DELETE request against the given path. v2 answers with 204 and no body.
func (c *Client) Delete(ctx context.Context, path string, options map[string]string) (io.Reader, error) {
	return c.do(ctx, http.MethodDelete, path, options, nil)
}

// do issues one request. The return values are named so that a failure to
// close the response body still reaches the caller.
func (c *Client) do(ctx context.Context, method, path string, options map[string]string, body any) (_ io.Reader, err error) {
	// JoinPath keeps the /v2 prefix on the base URL; assigning to u.Path would
	// drop it and silently send every request to the wrong place.
	u := c.Base.JoinPath(path)

	query := u.Query()
	for k, v := range options {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("could not marshal body: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w", method, u.Redacted(), err)
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

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	// v2 reports failures with a 4xx or 5xx status rather than an error
	// embedded in a 200, and answers writes with 201 or 204.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errResp := &ErrorResponse{StatusCode: resp.StatusCode, RawBody: buf.Bytes()}
		if err := json.Unmarshal(buf.Bytes(), errResp); err != nil {
			// Not the documented error body: a proxy or gateway may have
			// answered instead, so pass along whatever it said.
			if raw := strings.TrimSpace(buf.String()); raw != "" {
				return nil, fmt.Errorf("%s: %s", resp.Status, raw)
			}

			return nil, fmt.Errorf("%s", resp.Status)
		}

		// Wrap even when the body carried no message of its own, so that
		// errors.As still reaches the status code and the raw body.
		prefix := resp.Status
		if errResp.Error() != "" {
			prefix += ": "
		}

		return nil, fmt.Errorf("%s%w", prefix, errResp)
	}

	return &buf, nil
}

// ParseCurrency converts a string amount and currency code into a money.Money
// struct. The amount is parsed as an exact decimal rather than a float so that
// the four decimal places the v2 API returns do not pick up rounding error, and
// is scaled to the number of minor units the currency actually uses.
func ParseCurrency(amount, currency string) (*money.Money, error) {
	// go-money falls back to two minor units for codes it does not know, so
	// match that rather than dereferencing a nil currency.
	fraction := 2
	if c := money.GetCurrency(currency); c != nil {
		fraction = c.Fraction
	}

	units, err := parseDecimal(amount, fraction)
	if err != nil {
		return nil, err
	}

	return money.New(units, currency), nil
}

// parseDecimal converts a decimal string into an integer number of minor units,
// rounding half away from zero when the string carries more precision than the
// currency has room for.
func parseDecimal(amount string, fraction int) (int64, error) {
	s := strings.TrimSpace(amount)
	if s == "" {
		return 0, fmt.Errorf("%q is not a valid amount", amount)
	}

	neg := false
	switch s[0] {
	case '-':
		neg, s = true, s[1:]
	case '+':
		s = s[1:]
	}

	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" && frac == "" {
		return 0, fmt.Errorf("%q is not a valid amount", amount)
	}

	digits := whole + frac
	if digits == "" || strings.ContainsFunc(digits, func(r rune) bool { return r < '0' || r > '9' }) {
		return 0, fmt.Errorf("%q is not a valid amount", amount)
	}

	// Pad or trim the fractional digits to the currency's precision, rounding
	// on the first digit we drop.
	roundUp := false
	switch {
	case len(frac) < fraction:
		digits += strings.Repeat("0", fraction-len(frac))
	case len(frac) > fraction:
		cut := len(digits) - (len(frac) - fraction)
		roundUp = digits[cut] >= '5'
		digits = digits[:cut]
	}

	units, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a valid amount: %w", amount, err)
	}

	if roundUp {
		if units == math.MaxInt64 {
			return 0, fmt.Errorf("%q overflows int64", amount)
		}
		units++
	}

	if neg {
		units = -units
	}

	return units, nil
}
