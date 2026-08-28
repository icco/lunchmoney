package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// Cryptocurrency is one symbol supported for manual crypto tracking.
type Cryptocurrency struct {
	ID          int64  `json:"id"`
	CoingeckoID string `json:"coingecko_id"`
	Symbol      string `json:"symbol"`
	FullName    string `json:"full_name"`
}

type cryptocurrenciesResponse struct {
	Cryptocurrencies []*Cryptocurrency `json:"cryptocurrencies"`
}

// CreateCryptocurrency is the request body for adding a supported cryptocurrency.
type CreateCryptocurrency struct {
	CoingeckoURL string `json:"coingecko_url" validate:"required,min=1,max=200"`
}

// GetCryptocurrencies returns all symbols supported for manual crypto balances.
func (c *Client) GetCryptocurrencies(ctx context.Context) ([]*Cryptocurrency, error) {
	body, err := c.Get(ctx, "/cryptocurrencies", nil)
	if err != nil {
		return nil, fmt.Errorf("get cryptocurrencies: %w", err)
	}

	resp := &cryptocurrenciesResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.Cryptocurrencies, nil
}

// CreateCryptocurrency adds a supported cryptocurrency by CoinGecko URL.
func (c *Client) CreateCryptocurrency(ctx context.Context, req *CreateCryptocurrency) (*Cryptocurrency, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, req); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, "/cryptocurrencies", req)
	if err != nil {
		return nil, fmt.Errorf("create cryptocurrency: %w", err)
	}

	resp := &Cryptocurrency{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// CryptoManual is one manually managed crypto balance.
type CryptoManual struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	DisplayName      *string    `json:"display_name"`
	InstitutionName  *string    `json:"institution_name"`
	Balance          string     `json:"balance"`
	Symbol           string     `json:"symbol"`
	CoingeckoID      *string    `json:"coingecko_id"`
	ToBase           *float64   `json:"to_base"`
	BalanceAsOf      *Timestamp `json:"balance_as_of"`
	ExchangeRateAsOf *Timestamp `json:"exchange_rate_as_of"`
	CreatedByName    *string    `json:"created_by_name"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ParsedAmount converts the balance and symbol's currency into money.Money.
func (a *CryptoManual) ParsedAmount(primaryCurrency string) (*money.Money, error) {
	return ParseCurrency(a.Balance, primaryCurrency)
}

type cryptoManualResponse struct {
	CryptoManual []*CryptoManual `json:"crypto_manual"`
}

// CreateCryptoManual is the request body for creating a manual crypto balance.
type CreateCryptoManual struct {
	Name            string  `json:"name" validate:"required,min=1,max=45"`
	DisplayName     *string `json:"display_name,omitempty" validate:"omitnil,min=1,max=45"`
	InstitutionName *string `json:"institution_name,omitempty" validate:"omitnil,min=1,max=50"`
	Balance         string  `json:"balance" validate:"required"`
	Symbol          string  `json:"symbol" validate:"required,min=1,max=25"`
}

// UpdateCryptoManual is the request body for updating a manual crypto balance.
type UpdateCryptoManual struct {
	ID               *int64     `json:"id,omitempty"`
	Name             *string    `json:"name,omitempty" validate:"omitnil,min=1,max=45"`
	DisplayName      *string    `json:"display_name,omitempty" validate:"omitnil,min=1,max=45"`
	InstitutionName  *string    `json:"institution_name,omitempty" validate:"omitnil,min=1,max=50"`
	Balance          *string    `json:"balance,omitempty"`
	Symbol           *string    `json:"symbol,omitempty"`
	CoingeckoID      *string    `json:"coingecko_id,omitempty"`
	ToBase           *float64   `json:"to_base,omitempty"`
	BalanceAsOf      *Timestamp `json:"balance_as_of,omitempty"`
	ExchangeRateAsOf *Timestamp `json:"exchange_rate_as_of,omitempty"`
	CreatedByName    *string    `json:"created_by_name,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}

// GetCryptoManual returns all manual crypto balances.
func (c *Client) GetCryptoManual(ctx context.Context) ([]*CryptoManual, error) {
	body, err := c.Get(ctx, "/crypto/manual", nil)
	if err != nil {
		return nil, fmt.Errorf("get manual crypto balances: %w", err)
	}

	resp := &cryptoManualResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.CryptoManual, nil
}

// GetCryptoManualAccount returns one manual crypto balance by ID.
func (c *Client) GetCryptoManualAccount(ctx context.Context, id int64) (*CryptoManual, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/crypto/manual/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get manual crypto balance %d: %w", id, err)
	}

	resp := &CryptoManual{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// CreateCryptoManual creates a manual crypto balance and returns it.
func (c *Client) CreateCryptoManual(ctx context.Context, account *CreateCryptoManual) (*CryptoManual, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, account); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, "/crypto/manual", account)
	if err != nil {
		return nil, fmt.Errorf("create manual crypto balance: %w", err)
	}

	resp := &CryptoManual{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// UpdateCryptoManual updates one manual crypto balance and returns it.
func (c *Client) UpdateCryptoManual(ctx context.Context, id int64, account *UpdateCryptoManual) (*CryptoManual, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, account); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, fmt.Sprintf("/crypto/manual/%d", id), account)
	if err != nil {
		return nil, fmt.Errorf("update manual crypto balance %d: %w", id, err)
	}

	resp := &CryptoManual{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// DeleteCryptoManualOptions configures whether account history is retained.
type DeleteCryptoManualOptions struct {
	KeepHistory *bool
}

// ToMap turns options into query parameters.
func (o *DeleteCryptoManualOptions) ToMap() (map[string]string, error) {
	ret := map[string]string{}
	if o.KeepHistory != nil {
		ret["keep_history"] = strconv.FormatBool(*o.KeepHistory)
	}

	return ret, nil
}

// DeleteCryptoManual deletes one manual crypto balance.
func (c *Client) DeleteCryptoManual(ctx context.Context, id int64, opts *DeleteCryptoManualOptions) error {
	options := map[string]string{}
	if opts != nil {
		maps, err := opts.ToMap()
		if err != nil {
			return fmt.Errorf("convert options to map: %w", err)
		}
		options = maps
	}

	if _, err := c.Delete(ctx, fmt.Sprintf("/crypto/manual/%d", id), options); err != nil {
		return fmt.Errorf("delete manual crypto balance %d: %w", id, err)
	}

	return nil
}

// CryptoSyncedBalance is one symbol balance nested under a synced crypto account.
type CryptoSyncedBalance struct {
	Name             string     `json:"name"`
	DisplayName      *string    `json:"display_name"`
	Balance          string     `json:"balance"`
	Symbol           string     `json:"symbol"`
	CoingeckoID      *string    `json:"coingecko_id"`
	ToBase           *float64   `json:"to_base"`
	BalanceAsOf      *Timestamp `json:"balance_as_of"`
	ExchangeRateAsOf *Timestamp `json:"exchange_rate_as_of"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ParsedAmount converts the converted balance value to money.Money.
func (a *CryptoSyncedBalance) ParsedAmount(primaryCurrency string) (*money.Money, error) {
	return ParseCurrency(a.Balance, primaryCurrency)
}

// CryptoSyncedAccount is one synced crypto connection and its balances.
type CryptoSyncedAccount struct {
	ID            int64                  `json:"id"`
	Provider      string                 `json:"provider"`
	Status        string                 `json:"status"`
	CreatedByName *string                `json:"created_by_name"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	LastImport    *Timestamp             `json:"last_import"`
	DisplayName   *string                `json:"display_name"`
	Balances      []*CryptoSyncedBalance `json:"balances"`
}

type cryptoSyncedResponse struct {
	CryptoSynced []*CryptoSyncedAccount `json:"crypto_synced"`
}

// GetCryptoSynced returns all synced crypto accounts.
func (c *Client) GetCryptoSynced(ctx context.Context) ([]*CryptoSyncedAccount, error) {
	body, err := c.Get(ctx, "/crypto/synced", nil)
	if err != nil {
		return nil, fmt.Errorf("get synced crypto accounts: %w", err)
	}

	resp := &cryptoSyncedResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.CryptoSynced, nil
}

// GetCryptoSyncedAccount returns one synced crypto account by ID.
func (c *Client) GetCryptoSyncedAccount(ctx context.Context, id int64) (*CryptoSyncedAccount, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/crypto/synced/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get synced crypto account %d: %w", id, err)
	}

	resp := &CryptoSyncedAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// GetCryptoSyncedBalance returns one symbol balance from a synced crypto account.
func (c *Client) GetCryptoSyncedBalance(ctx context.Context, id int64, symbol string) (*CryptoSyncedBalance, error) {
	symbol = strings.TrimSpace(symbol)

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.VarCtx(ctx, symbol, "required,min=1,max=25"); err != nil {
		return nil, err
	}

	body, err := c.Get(ctx, fmt.Sprintf("/crypto/synced/%d/%s", id, symbol), nil)
	if err != nil {
		return nil, fmt.Errorf("get synced crypto balance %d/%s: %w", id, symbol, err)
	}

	resp := &CryptoSyncedBalance{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// RefreshCryptoSynced triggers a refresh and returns the updated synced account.
func (c *Client) RefreshCryptoSynced(ctx context.Context, id int64) (*CryptoSyncedAccount, error) {
	body, err := c.Post(ctx, fmt.Sprintf("/crypto/synced/%d/refresh", id), nil)
	if err != nil {
		return nil, fmt.Errorf("refresh synced crypto account %d: %w", id, err)
	}

	resp := &CryptoSyncedAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
