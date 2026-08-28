package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

var monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// BalanceHistoryFilters are optional month-range filters for balance history.
// StartMonth and EndMonth must be provided together.
type BalanceHistoryFilters struct {
	StartMonth *string
	EndMonth   *string
}

// ToMap converts filter values to query parameters.
func (f *BalanceHistoryFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}
	if f == nil {
		return ret, nil
	}

	if (f.StartMonth == nil) != (f.EndMonth == nil) {
		return nil, fmt.Errorf("start_month and end_month must either both be provided or both be omitted")
	}

	if f.StartMonth != nil {
		start := strings.TrimSpace(*f.StartMonth)
		end := strings.TrimSpace(*f.EndMonth)
		if !monthPattern.MatchString(start) {
			return nil, fmt.Errorf("start_month must be in YYYY-MM format")
		}
		if !monthPattern.MatchString(end) {
			return nil, fmt.Errorf("end_month must be in YYYY-MM format")
		}

		ret["start_month"] = start
		ret["end_month"] = end
	}

	return ret, nil
}

// BalanceHistorySource identifies an account source in balance history output.
type BalanceHistorySource struct {
	Type             string  `json:"type"`
	ManualAccountID  *int64  `json:"manual_account_id"`
	PlaidAccountID   *int64  `json:"plaid_account_id"`
	CryptoManualID   *int64  `json:"crypto_manual_id"`
	CryptoSyncedID   *int64  `json:"crypto_synced_id"`
	DeletedAccountID *int64  `json:"deleted_account_id"`
	Name             *string `json:"name"`
	InstitutionName  *string `json:"institution_name"`
	DisplayName      *string `json:"display_name"`
	AccountType      *string `json:"account_type"`
	Subtype          *string `json:"subtype"`
	Mask             *string `json:"mask"`
	Symbol           *string `json:"symbol"`
}

// BalanceHistoryEntry is one balance datapoint for a month.
type BalanceHistoryEntry struct {
	Type          string  `json:"type"`
	ID            *int64  `json:"id,omitempty"`
	Month         string  `json:"month"`
	Balance       string  `json:"balance"`
	Currency      string  `json:"currency"`
	ToBase        float64 `json:"to_base"`
	CryptoBalance *string `json:"crypto_balance"`
}

// ParsedAmount converts the balance and currency into money.Money.
func (e *BalanceHistoryEntry) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(e.Balance, e.Currency)
}

// BalanceHistoryAccount groups one source account and its monthly balances.
type BalanceHistoryAccount struct {
	Source   *BalanceHistorySource  `json:"source"`
	Balances []*BalanceHistoryEntry `json:"balances"`
}

type balanceHistoryResponse struct {
	BalanceHistory []*BalanceHistoryAccount `json:"balance_history"`
}

// BalanceHistoryAccountType names path values accepted by balance-history account endpoints.
type BalanceHistoryAccountType string

const (
	BalanceHistoryAccountTypeManual       BalanceHistoryAccountType = "manual"
	BalanceHistoryAccountTypePlaid        BalanceHistoryAccountType = "plaid"
	BalanceHistoryAccountTypeCryptoManual BalanceHistoryAccountType = "crypto_manual"
	BalanceHistoryAccountTypeDeleted      BalanceHistoryAccountType = "deleted"
)

func (t BalanceHistoryAccountType) validate(ctx context.Context) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.VarCtx(ctx, string(t), "required,oneof=manual plaid crypto_manual deleted")
}

// GetBalanceHistory returns balance history grouped by account source.
func (c *Client) GetBalanceHistory(ctx context.Context, filters *BalanceHistoryFilters) ([]*BalanceHistoryAccount, error) {
	options, err := filters.ToMap()
	if err != nil {
		return nil, fmt.Errorf("convert filters to map: %w", err)
	}

	body, err := c.Get(ctx, "/balance_history", options)
	if err != nil {
		return nil, fmt.Errorf("get balance history: %w", err)
	}

	resp := &balanceHistoryResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.BalanceHistory, nil
}

// GetBalanceHistoryForAccount returns balance history for one account source.
func (c *Client) GetBalanceHistoryForAccount(ctx context.Context, accountType BalanceHistoryAccountType, accountID int64, filters *BalanceHistoryFilters) ([]*BalanceHistoryAccount, error) {
	if err := accountType.validate(ctx); err != nil {
		return nil, err
	}

	options, err := filters.ToMap()
	if err != nil {
		return nil, fmt.Errorf("convert filters to map: %w", err)
	}

	path := fmt.Sprintf("/balance_history/%s/%d", accountType, accountID)
	body, err := c.Get(ctx, path, options)
	if err != nil {
		return nil, fmt.Errorf("get balance history for %s account %d: %w", accountType, accountID, err)
	}

	resp := &balanceHistoryResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.BalanceHistory, nil
}

// UpsertBalanceHistory is the request payload for balance history upserts.
type UpsertBalanceHistory struct {
	Balances []*UpsertBalanceHistoryItem `json:"balances" validate:"required,min=1,dive"`
}

// UpsertBalanceHistoryItem is one monthly history entry for an upsert.
type UpsertBalanceHistoryItem struct {
	ID            *int64   `json:"id,omitempty"`
	Month         string   `json:"month" validate:"required"`
	Balance       string   `json:"balance" validate:"required"`
	Symbol        *string  `json:"symbol,omitempty" validate:"omitnil,min=1,max=25"`
	CryptoBalance *string  `json:"crypto_balance,omitempty"`
	Currency      *string  `json:"currency,omitempty" validate:"omitnil,len=3"`
	ToBase        *float64 `json:"to_base,omitempty"`
}

func (u *UpsertBalanceHistory) validate(ctx context.Context) error {
	if u == nil {
		return fmt.Errorf("request is required")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, u); err != nil {
		return err
	}

	for i, item := range u.Balances {
		if item == nil {
			return fmt.Errorf("balances[%d] cannot be nil", i)
		}
		if !monthPattern.MatchString(strings.TrimSpace(item.Month)) {
			return fmt.Errorf("balances[%d].month must be in YYYY-MM format", i)
		}
	}

	return nil
}

// UpsertBalanceHistoryForAccount upserts one or more entries for one account source.
func (c *Client) UpsertBalanceHistoryForAccount(ctx context.Context, accountType BalanceHistoryAccountType, accountID int64, request *UpsertBalanceHistory) (*BalanceHistoryAccount, error) {
	if err := accountType.validate(ctx); err != nil {
		return nil, err
	}
	if err := request.validate(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/balance_history/%s/%d", accountType, accountID)
	body, err := c.Put(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("upsert balance history for %s account %d: %w", accountType, accountID, err)
	}

	resp := &BalanceHistoryAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// DeleteBalanceHistoryForAccount deletes all history entries for one account source.
func (c *Client) DeleteBalanceHistoryForAccount(ctx context.Context, accountType BalanceHistoryAccountType, accountID int64) error {
	if err := accountType.validate(ctx); err != nil {
		return err
	}

	path := fmt.Sprintf("/balance_history/%s/%d", accountType, accountID)
	if _, err := c.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("delete balance history for %s account %d: %w", accountType, accountID, err)
	}

	return nil
}

// GetBalanceHistoryForCryptoSynced returns balance history for one synced symbol stream.
func (c *Client) GetBalanceHistoryForCryptoSynced(ctx context.Context, accountID int64, symbol string, filters *BalanceHistoryFilters) ([]*BalanceHistoryAccount, error) {
	symbol = strings.TrimSpace(symbol)

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.VarCtx(ctx, symbol, "required,min=1,max=25"); err != nil {
		return nil, err
	}

	options, err := filters.ToMap()
	if err != nil {
		return nil, fmt.Errorf("convert filters to map: %w", err)
	}

	path := fmt.Sprintf("/balance_history/crypto_synced/%d/%s", accountID, symbol)
	body, err := c.Get(ctx, path, options)
	if err != nil {
		return nil, fmt.Errorf("get balance history for synced crypto %d/%s: %w", accountID, symbol, err)
	}

	resp := &balanceHistoryResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp.BalanceHistory, nil
}

// UpsertBalanceHistoryForCryptoSynced upserts one or more entries for a synced symbol stream.
func (c *Client) UpsertBalanceHistoryForCryptoSynced(ctx context.Context, accountID int64, symbol string, request *UpsertBalanceHistory) (*BalanceHistoryAccount, error) {
	symbol = strings.TrimSpace(symbol)

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.VarCtx(ctx, symbol, "required,min=1,max=25"); err != nil {
		return nil, err
	}
	if err := request.validate(ctx); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/balance_history/crypto_synced/%d/%s", accountID, symbol)
	body, err := c.Put(ctx, path, request)
	if err != nil {
		return nil, fmt.Errorf("upsert balance history for synced crypto %d/%s: %w", accountID, symbol, err)
	}

	resp := &BalanceHistoryAccount{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// DeleteBalanceHistoryForCryptoSynced deletes all entries for a synced symbol stream.
func (c *Client) DeleteBalanceHistoryForCryptoSynced(ctx context.Context, accountID int64, symbol string) error {
	symbol = strings.TrimSpace(symbol)

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.VarCtx(ctx, symbol, "required,min=1,max=25"); err != nil {
		return err
	}

	path := fmt.Sprintf("/balance_history/crypto_synced/%d/%s", accountID, symbol)
	if _, err := c.Delete(ctx, path, nil); err != nil {
		return fmt.Errorf("delete balance history for synced crypto %d/%s: %w", accountID, symbol, err)
	}

	return nil
}
