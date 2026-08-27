package lunchmoney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// PlaidAccountsResponse is a list plaid accounts response.
type PlaidAccountsResponse struct {
	PlaidAccounts []*PlaidAccount `json:"plaid_accounts"`
}

// PlaidAccount is a single LM Plaid account.
type PlaidAccount struct {
	ID              int64  `json:"id"`
	PlaidItemID     string `json:"plaid_item_id"`
	DateLinked      string `json:"date_linked"`
	LinkedByName    string `json:"linked_by_name"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Type            string `json:"type"`
	Subtype         string `json:"subtype"`
	Mask            string `json:"mask"`
	InstitutionName string `json:"institution_name"`
	// Status is one of active, inactive, closed, deactivated, not found,
	// not supported, relink, syncing, revoked or error.
	Status                        string     `json:"status"`
	AllowTransactionModifications bool       `json:"allow_transaction_modifications"`
	Limit                         *float64   `json:"limit"`
	Balance                       string     `json:"balance"`
	Currency                      string     `json:"currency"`
	ToBase                        float64    `json:"to_base"` // the balance converted to the user's primary currency
	BalanceLastUpdate             *time.Time `json:"balance_last_update"`
	ImportStartDate               string     `json:"import_start_date"`
	LastImport                    *time.Time `json:"last_import"`
	LastFetch                     *time.Time `json:"last_fetch"`
	PlaidLastSuccessfulUpdate     *time.Time `json:"plaid_last_successful_update"`
}

// ParsedAmount converts the account's balance and currency into a money.Money.
func (p *PlaidAccount) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(p.Balance, p.Currency)
}

// GetPlaidAccounts retrieves all Plaid-connected accounts.
func (c *Client) GetPlaidAccounts(ctx context.Context) ([]*PlaidAccount, error) {
	body, err := c.Get(ctx, "/plaid_accounts", nil)
	if err != nil {
		return nil, fmt.Errorf("get plaid accounts: %w", err)
	}

	resp := &PlaidAccountsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.PlaidAccounts, nil
}

// GetPlaidAccount retrieves a single Plaid-connected account by its ID.
func (c *Client) GetPlaidAccount(ctx context.Context, id int64) (*PlaidAccount, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/plaid_accounts/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get plaid account %d: %w", id, err)
	}

	resp := &PlaidAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// PlaidFetchFilters narrows what TriggerPlaidFetch asks Plaid for. All fields
// are optional; a nil filter fetches every eligible account.
type PlaidFetchFilters struct {
	// StartDate and EndDate bound the transactions to fetch. The API requires
	// both or neither, so required_with comes before omitnil: omitnil
	// short-circuits the rest of the tag when the field is nil, which is
	// exactly the case the pairing rule has to catch.
	StartDate *string `validate:"required_with=EndDate,omitnil,datetime=2006-01-02"`
	EndDate   *string `validate:"required_with=StartDate,omitnil,datetime=2006-01-02"`

	// ID limits the fetch to a single Plaid account.
	ID *int64
}

// ToMap converts the filters to a string map to be sent as query parameters.
// If a field is nil, it will not be included in the map.
func (r *PlaidFetchFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}

	strs := map[string]*string{
		queryStartDate: r.StartDate,
		queryEndDate:   r.EndDate,
	}
	for k, v := range strs {
		if v != nil {
			ret[k] = *v
		}
	}

	if r.ID != nil {
		ret["id"] = strconv.FormatInt(*r.ID, 10)
	}

	return ret, nil
}

// IsTooEarly reports whether err is a 425 Too Early response, which the API
// returns when a Plaid fetch was already triggered within the last minute.
func IsTooEarly(err error) bool {
	var resp *ErrorResponse

	return errors.As(err, &resp) && resp.StatusCode == http.StatusTooEarly
}

// TriggerPlaidFetch queues a background fetch from Plaid. It answers 425 if one
// was triggered in the last minute; check with IsTooEarly.
func (c *Client) TriggerPlaidFetch(ctx context.Context, filters *PlaidFetchFilters) ([]*PlaidAccount, error) {
	options := map[string]string{}
	if filters != nil {
		validate := validator.New(validator.WithRequiredStructEnabled())
		if err := validate.StructCtx(ctx, filters); err != nil {
			return nil, err
		}

		maps, err := filters.ToMap()
		if err != nil {
			return nil, fmt.Errorf("convert filters to map: %w", err)
		}
		options = maps
	}

	// The filters are query parameters on a POST, which Client.Post cannot
	// express.
	body, err := c.do(ctx, http.MethodPost, "/plaid_accounts/fetch", options, nil)
	if err != nil {
		return nil, fmt.Errorf("trigger plaid fetch: %w", err)
	}

	// The spec documents no body on the 202, so an empty one is not an error.
	resp := &PlaidAccountsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.PlaidAccounts, nil
}
