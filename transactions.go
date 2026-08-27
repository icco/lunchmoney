package lunchmoney

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// Transaction status values. v1's "cleared" and "uncleared" became "reviewed"
// and "unreviewed"; its "pending" and "recurring" statuses are gone, and
// pending transactions are now flagged by the IsPending field.
const (
	StatusReviewed      = "reviewed"
	StatusUnreviewed    = "unreviewed"
	StatusDeletePending = "delete_pending"
)

// TransactionsResponse is the response we get from requesting transactions.
type TransactionsResponse struct {
	Transactions []*Transaction `json:"transactions"`
	HasMore      bool           `json:"has_more"`
}

// Transaction is a single LM transaction.
//
// Amounts no longer follow the account's debit_as_negative preference: a
// positive amount is always a debit and a negative amount always a credit.
//
// v2 does not hydrate related records onto the transaction. The category and
// account name fields v1 returned are gone; look them up with GetCategory,
// GetManualAccount or GetPlaidAccount instead.
type Transaction struct {
	ID              int64   `json:"id"`
	Date            string  `json:"date"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	ToBase          float64 `json:"to_base"`
	RecurringID     *int64  `json:"recurring_id"`
	Payee           string  `json:"payee"`
	OriginalName    string  `json:"original_name"`
	CategoryID      *int64  `json:"category_id"`
	PlaidAccountID  *int64  `json:"plaid_account_id"`
	ManualAccountID *int64  `json:"manual_account_id"`
	ExternalID      string  `json:"external_id"`
	TagIDs          []int64 `json:"tag_ids"`
	Notes           string  `json:"notes"`
	// Status is one of StatusReviewed, StatusUnreviewed or StatusDeletePending.
	Status        string    `json:"status"`
	IsPending     bool      `json:"is_pending"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsSplitParent bool      `json:"is_split_parent"`
	SplitParentID *int64    `json:"split_parent_id"`
	IsGroupParent bool      `json:"is_group_parent"`
	GroupParentID *int64    `json:"group_parent_id"`
	// Children is populated for split parents and transaction groups when the
	// IncludeChildren filter is set.
	Children []*Transaction `json:"children,omitempty"`
	// PlaidMetadata and CustomMetadata are only present when the
	// IncludeMetadata filter is set. Their schemas are variable.
	PlaidMetadata  map[string]any `json:"plaid_metadata,omitempty"`
	CustomMetadata map[string]any `json:"custom_metadata,omitempty"`
	// Files is only present when the IncludeFiles filter is set.
	Files []TransactionAttachment `json:"files,omitempty"`
	// Source is one of api, csv, manual, merge, plaid, recurring, rule, split
	// or user.
	Source string `json:"source"`
}

// TransactionAttachment describes a file attached to a transaction.
type TransactionAttachment struct {
	ID         int64     `json:"id"`
	UploadedBy int64     `json:"uploaded_by"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Size       int64     `json:"size"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

// ParsedAmount converts the transaction's amount and currency into a money.Money.
func (t *Transaction) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(t.Amount, t.Currency)
}

// TransactionFilters are options to pass into the request for transactions.
// The v1 DebitAsNegative option is gone, and AssetID is now ManualAccountID.
type TransactionFilters struct {
	StartDate    *string `validate:"omitnil,datetime=2006-01-02"`
	EndDate      *string `validate:"omitnil,datetime=2006-01-02"`
	CreatedSince *string
	UpdatedSince *string

	TagID           *int64
	RecurringID     *int64
	PlaidAccountID  *int64
	ManualAccountID *int64
	CategoryID      *int64

	IsGroupParent *bool
	Status        *string `validate:"omitnil,oneof=reviewed unreviewed delete_pending"`
	IsPending     *bool

	IncludePending       *bool
	IncludeMetadata      *bool
	IncludeSplitParents  *bool
	IncludeGroupChildren *bool
	IncludeChildren      *bool
	IncludeFiles         *bool

	Limit  *int64 `validate:"omitnil,min=1,max=2000"`
	Offset *int64
}

// ToMap converts the filters to a string map to be sent with the request as
// GET parameters. If the field is nil, it will not be included in the map.
func (r *TransactionFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{}

	strs := map[string]*string{
		queryStartDate:  r.StartDate,
		queryEndDate:    r.EndDate,
		"created_since": r.CreatedSince,
		"updated_since": r.UpdatedSince,
		"status":        r.Status,
	}
	for k, v := range strs {
		if v != nil {
			ret[k] = *v
		}
	}

	ints := map[string]*int64{
		"tag_id":            r.TagID,
		"recurring_id":      r.RecurringID,
		"plaid_account_id":  r.PlaidAccountID,
		"manual_account_id": r.ManualAccountID,
		"category_id":       r.CategoryID,
		"limit":             r.Limit,
		"offset":            r.Offset,
	}
	for k, v := range ints {
		if v != nil {
			ret[k] = strconv.FormatInt(*v, 10)
		}
	}

	bools := map[string]*bool{
		"is_group_parent":        r.IsGroupParent,
		"is_pending":             r.IsPending,
		"include_pending":        r.IncludePending,
		"include_metadata":       r.IncludeMetadata,
		"include_split_parents":  r.IncludeSplitParents,
		"include_group_children": r.IncludeGroupChildren,
		"include_children":       r.IncludeChildren,
		"include_files":          r.IncludeFiles,
	}
	for k, v := range bools {
		if v != nil {
			ret[k] = strconv.FormatBool(*v)
		}
	}

	return ret, nil
}

// GetTransactions retrieves transactions matching filters, plus a flag saying
// whether more remain beyond this page.
func (c *Client) GetTransactions(ctx context.Context, filters *TransactionFilters) (*TransactionsResponse, error) {
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

	body, err := c.Get(ctx, "/transactions", options)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	resp := &TransactionsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// GetTransaction retrieves a single transaction by its ID.
func (c *Client) GetTransaction(ctx context.Context, id int64) (*Transaction, error) {
	body, err := c.Get(ctx, fmt.Sprintf("/transactions/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("get transaction %d: %w", id, err)
	}

	resp := &Transaction{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// InsertTransactionsRequest creates one or more transactions.
type InsertTransactionsRequest struct {
	Transactions      []InsertTransaction `json:"transactions" validate:"min=1,max=500,dive"`
	ApplyRules        bool                `json:"apply_rules,omitempty"`
	SkipDuplicates    bool                `json:"skip_duplicates,omitempty"`
	SkipBalanceUpdate bool                `json:"skip_balance_update,omitempty"`
}

// InsertTransaction is a transaction to create. Date and Amount are required.
// Tags must already exist: v2 takes IDs and will not create tags inline.
type InsertTransaction struct {
	Date            string         `json:"date" validate:"datetime=2006-01-02"`
	Amount          string         `json:"amount" validate:"required"`
	Currency        string         `json:"currency,omitempty"`
	Payee           string         `json:"payee,omitempty"`
	OriginalName    string         `json:"original_name,omitempty"`
	CategoryID      *int64         `json:"category_id,omitempty"`
	Notes           string         `json:"notes,omitempty"`
	ManualAccountID *int64         `json:"manual_account_id,omitempty"`
	PlaidAccountID  *int64         `json:"plaid_account_id,omitempty"`
	RecurringID     *int64         `json:"recurring_id,omitempty"`
	Status          string         `json:"status,omitempty" validate:"omitempty,oneof=reviewed unreviewed"`
	TagIDs          []int64        `json:"tag_ids,omitempty"`
	ExternalID      string         `json:"external_id,omitempty" validate:"max=75"`
	CustomMetadata  map[string]any `json:"custom_metadata,omitempty"`
}

// InsertTransactionsResponse holds the results of an InsertTransactions call.
//
// Unlike v1, a duplicate no longer fails the whole request: the transactions
// that were accepted come back in Transactions, and the ones that were not are
// reported in SkippedDuplicates.
type InsertTransactionsResponse struct {
	Transactions      []*Transaction     `json:"transactions"`
	SkippedDuplicates []SkippedDuplicate `json:"skipped_duplicates"`
}

// SkippedDuplicate describes a transaction the API declined to insert because
// it matched an existing one.
type SkippedDuplicate struct {
	// Reason is either "duplicate_external_id" or "duplicate_payee_amount_date".
	Reason                  string            `json:"reason"`
	RequestTransactionIndex int64             `json:"request_transactions_index"`
	ExistingTransactionID   int64             `json:"existing_transaction_id"`
	RequestTransaction      InsertTransaction `json:"request_transaction"`
}

// InsertTransactions creates new transactions.
func (c *Client) InsertTransactions(ctx context.Context, itReq InsertTransactionsRequest) (*InsertTransactionsResponse, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, itReq); err != nil {
		return nil, err
	}

	body, err := c.Post(ctx, "/transactions", itReq)
	if err != nil {
		return nil, fmt.Errorf("insert transactions: %w", err)
	}

	resp := &InsertTransactionsResponse{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("insert response decode error: %w", err)
	}

	return resp, nil
}

// UpdateTransaction holds the updatable fields of a transaction. Only non-nil
// fields are sent.
type UpdateTransaction struct {
	Date            *string `json:"date,omitempty" validate:"omitnil,datetime=2006-01-02"`
	Amount          *string `json:"amount,omitempty"`
	Currency        *string `json:"currency,omitempty"`
	Payee           *string `json:"payee,omitempty"`
	CategoryID      *int64  `json:"category_id,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	ManualAccountID *int64  `json:"manual_account_id,omitempty"`
	PlaidAccountID  *int64  `json:"plaid_account_id,omitempty"`
	RecurringID     *int64  `json:"recurring_id,omitempty"`
	Status          *string `json:"status,omitempty" validate:"omitnil,oneof=reviewed unreviewed"`
	// TagIDs replaces the transaction's tags; AdditionalTagIDs adds to them.
	TagIDs           *[]int64        `json:"tag_ids,omitempty"`
	AdditionalTagIDs *[]int64        `json:"additional_tag_ids,omitempty"`
	ExternalID       *string         `json:"external_id,omitempty" validate:"omitnil,max=75"`
	CustomMetadata   *map[string]any `json:"custom_metadata,omitempty"`
}

// UpdateTransaction modifies an existing transaction with the specified ID and
// returns the updated transaction.
//
// Splitting moved out of this call in v2. The split payload v1 accepted here
// is now POST /transactions/split/{id}, which this library does not wrap yet.
func (c *Client) UpdateTransaction(ctx context.Context, id int64, ut *UpdateTransaction) (*Transaction, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, ut); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, fmt.Sprintf("/transactions/%d", id), ut)
	if err != nil {
		return nil, fmt.Errorf("update transaction %d: %w", id, err)
	}

	resp := &Transaction{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}
