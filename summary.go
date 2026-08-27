package lunchmoney

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Rhymond/go-money"
	"github.com/go-playground/validator/v10"
)

// BudgetSummary is the budget and activity rollup for a date range. It replaces
// the per-category, per-month Budget list v1 returned from /budgets.
type BudgetSummary struct {
	// Aligned reports whether the requested range lines up with the user's
	// configured budget periods. When false, the numbers cover a range that
	// spans or splits budget periods.
	Aligned      bool               `json:"aligned"`
	Categories   []*SummaryCategory `json:"categories"`
	Totals       *SummaryTotals     `json:"totals,omitempty"`
	RolloverPool *RolloverPool      `json:"rollover_pool,omitempty"`
}

// SummaryCategory is one category's budget and activity within the range.
type SummaryCategory struct {
	CategoryID  int64                `json:"category_id"`
	Totals      SummaryCategoryTotal `json:"totals"`
	Occurrences []SummaryOccurrence  `json:"occurrences,omitempty"`
}

// SummaryCategoryTotal holds a category's amounts across the whole range, in
// the user's primary currency.
type SummaryCategoryTotal struct {
	OtherActivity      float64  `json:"other_activity"`
	RecurringActivity  float64  `json:"recurring_activity"`
	Budgeted           *float64 `json:"budgeted"`
	Available          *float64 `json:"available"`
	RecurringRemaining float64  `json:"recurring_remaining"`
	RecurringExpected  float64  `json:"recurring_expected"`
}

// SummaryOccurrence is a category's amounts for a single budget period within
// the requested range.
type SummaryOccurrence struct {
	InRange           bool     `json:"in_range"`
	StartDate         string   `json:"start_date"`
	EndDate           string   `json:"end_date"`
	OtherActivity     float64  `json:"other_activity"`
	RecurringActivity float64  `json:"recurring_activity"`
	Budgeted          *float64 `json:"budgeted"`
	BudgetedAmount    string   `json:"budgeted_amount"`
	BudgetedCurrency  string   `json:"budgeted_currency"`
	Notes             string   `json:"notes"`
}

// ParsedAmount converts the budgeted amount and currency into a money.Money.
func (o *SummaryOccurrence) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(o.BudgetedAmount, o.BudgetedCurrency)
}

// SummaryTotals splits the range's activity into money coming in and going out.
// It is only present when the IncludeTotals filter is set.
type SummaryTotals struct {
	Inflow  SummaryTotalsBreakdown `json:"inflow"`
	Outflow SummaryTotalsBreakdown `json:"outflow"`
}

// SummaryTotalsBreakdown is one direction's activity across the range.
type SummaryTotalsBreakdown struct {
	OtherActivity          float64 `json:"other_activity"`
	RecurringActivity      float64 `json:"recurring_activity"`
	RecurringRemaining     float64 `json:"recurring_remaining"`
	RecurringExpected      float64 `json:"recurring_expected"`
	Uncategorized          float64 `json:"uncategorized"`
	UncategorizedCount     int64   `json:"uncategorized_count"`
	UncategorizedRecurring float64 `json:"uncategorized_recurring"`
}

// RolloverPool is the left-to-budget pool and the adjustments made to it. It is
// only present when the IncludeRolloverPool filter is set.
type RolloverPool struct {
	BudgetedToBase float64              `json:"budgeted_to_base"`
	AllAdjustments []RolloverAdjustment `json:"all_adjustments"`
}

// RolloverAdjustment is a single change to the rollover pool.
type RolloverAdjustment struct {
	InRange  bool    `json:"in_range"`
	Date     string  `json:"date"`
	Amount   string  `json:"amount"`
	Currency string  `json:"currency"`
	ToBase   float64 `json:"to_base"`
}

// ParsedAmount converts the adjustment's amount and currency into a money.Money.
func (a *RolloverAdjustment) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(a.Amount, a.Currency)
}

// BudgetFilters are options to pass into the request for a budget summary.
// Both dates are required.
type BudgetFilters struct {
	StartDate string `validate:"datetime=2006-01-02,required"`
	EndDate   string `validate:"datetime=2006-01-02,required"`

	IncludeExcludeFromBudgets *bool
	IncludeOccurrences        *bool
	IncludePastBudgetDates    *bool
	IncludeTotals             *bool
	IncludeRolloverPool       *bool
}

// ToMap converts the budget filters to a string map to be sent with the request
// as GET parameters. Unset optional fields are omitted.
func (r *BudgetFilters) ToMap() (map[string]string, error) {
	ret := map[string]string{
		queryStartDate: r.StartDate,
		queryEndDate:   r.EndDate,
	}

	bools := map[string]*bool{
		"include_exclude_from_budgets": r.IncludeExcludeFromBudgets,
		"include_occurrences":          r.IncludeOccurrences,
		"include_past_budget_dates":    r.IncludePastBudgetDates,
		"include_totals":               r.IncludeTotals,
		"include_rollover_pool":        r.IncludeRolloverPool,
	}
	for k, v := range bools {
		if v != nil {
			ret[k] = strconv.FormatBool(*v)
		}
	}

	return ret, nil
}

// GetBudgetSummary returns the budget and activity summary for a period. It
// replaces v1's GetBudgets.
func (c *Client) GetBudgetSummary(ctx context.Context, filters *BudgetFilters) (*BudgetSummary, error) {
	if filters == nil {
		return nil, fmt.Errorf("filters with a start and end date are required")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, filters); err != nil {
		return nil, err
	}

	options, err := filters.ToMap()
	if err != nil {
		return nil, fmt.Errorf("convert filters to map: %w", err)
	}

	body, err := c.Get(ctx, "/summary", options)
	if err != nil {
		return nil, fmt.Errorf("get budget summary: %w", err)
	}

	resp := &BudgetSummary{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// BudgetSettings describes how the user's budget periods are configured.
type BudgetSettings struct {
	// PeriodGranularity is one of day, week, month, year or "twice a month".
	PeriodGranularity string  `json:"budget_period_granularity"`
	PeriodQuantity    float64 `json:"budget_period_quantity"`
	PeriodAnchorDate  string  `json:"budget_period_anchor_date"`
	HideNoActivity    bool    `json:"budget_hide_no_activity"`
	UseLastDayOfMonth bool    `json:"budget_use_last_day_of_month"`
	// IncomeOption is one of max, budgeted or activity.
	IncomeOption         string `json:"budget_income_option"`
	RolloverLeftToBudget bool   `json:"budget_rollover_left_to_budget"`
}

// GetBudgetSettings returns the user's budget period configuration.
func (c *Client) GetBudgetSettings(ctx context.Context) (*BudgetSettings, error) {
	body, err := c.Get(ctx, "/budgets/settings", nil)
	if err != nil {
		return nil, fmt.Errorf("get budget settings: %w", err)
	}

	resp := &BudgetSettings{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// Budget is a single category's budget for one period.
type Budget struct {
	CategoryID int64   `json:"category_id"`
	StartDate  string  `json:"start_date"`
	Amount     string  `json:"amount"`
	Currency   string  `json:"currency"`
	ToBase     float64 `json:"to_base"`
	Notes      string  `json:"notes"`
}

// UnmarshalJSON decodes a budget, accepting an amount given as a JSON number
// as well as the string the schema documents. The upsert response is the one
// place the API's own example shows an unquoted amount, and a failed decode
// there would report an error for a budget that was in fact written.
func (b *Budget) UnmarshalJSON(data []byte) error {
	type budget Budget

	aux := struct {
		Amount json.RawMessage `json:"amount"`
		*budget
	}{budget: (*budget)(b)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if len(aux.Amount) == 0 || string(aux.Amount) == "null" {
		return nil
	}

	if err := json.Unmarshal(aux.Amount, &b.Amount); err == nil {
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(aux.Amount, &number); err != nil {
		return fmt.Errorf("%s is not a valid amount: %w", aux.Amount, err)
	}

	b.Amount = number.String()

	return nil
}

// ParsedAmount converts the budgeted amount and currency into a money.Money.
func (b *Budget) ParsedAmount() (*money.Money, error) {
	return ParseCurrency(b.Amount, b.Currency)
}

// UpsertBudget is a budget to set. StartDate has to be a period start for the
// account, which GetBudgetSettings describes.
type UpsertBudget struct {
	StartDate  string `json:"start_date" validate:"required,datetime=2006-01-02"`
	CategoryID int64  `json:"category_id" validate:"required"`
	Amount     string `json:"amount" validate:"required"`

	// Currency defaults to the account's primary currency when empty.
	Currency string `json:"currency,omitempty"`

	// Notes is a pointer so that an empty string clears the stored notes
	// rather than being indistinguishable from leaving them alone.
	Notes *string `json:"notes,omitempty" validate:"omitnil,max=350"`
}

// UpsertBudget sets the budget for a category and period, creating it if there
// is none, and returns it.
func (c *Client) UpsertBudget(ctx context.Context, budget *UpsertBudget) (*Budget, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.StructCtx(ctx, budget); err != nil {
		return nil, err
	}

	body, err := c.Put(ctx, "/budgets", budget)
	if err != nil {
		return nil, fmt.Errorf("upsert budget: %w", budgetInvalidPeriod(err))
	}

	resp := &Budget{}
	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return resp, nil
}

// DeleteBudget removes the budget for a category and period. Removing a budget
// that is not set succeeds.
func (c *Client) DeleteBudget(ctx context.Context, categoryID int64, startDate string) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.VarCtx(ctx, startDate, "required,datetime=2006-01-02"); err != nil {
		return err
	}

	options := map[string]string{
		"category_id":  strconv.FormatInt(categoryID, 10),
		queryStartDate: startDate,
	}

	if _, err := c.Delete(ctx, "/budgets", options); err != nil {
		return fmt.Errorf("delete budget: %w", budgetInvalidPeriod(err))
	}

	return nil
}

// BudgetInvalidPeriodError is the 400 the API answers with when a start date is
// not a period start for the account. The budget calls return one so the valid
// dates on either side of the rejected one can be reached with errors.As.
type BudgetInvalidPeriodError struct {
	Message            string `json:"message"`
	ErrMsg             string `json:"errMsg"`
	RequestedStartDate string `json:"requested_start_date"`

	// PreviousValidStartDate and NextValidStartDate are empty when the
	// requested date has no valid period start before or after it.
	PreviousValidStartDate string `json:"previous_valid_start_date"`
	NextValidStartDate     string `json:"next_valid_start_date"`

	// Err is the API error the period details arrived as.
	Err error `json:"-"`
}

func (e *BudgetInvalidPeriodError) Error() string {
	return fmt.Sprintf("%s (requested %q, previous valid %q, next valid %q)", e.ErrMsg, e.RequestedStartDate, e.PreviousValidStartDate, e.NextValidStartDate)
}

func (e *BudgetInvalidPeriodError) Unwrap() error { return e.Err }

// budgetInvalidPeriod replaces a 400 that carries period details with them, and
// passes any other error, a plain validation 400 included, through untouched.
func budgetInvalidPeriod(err error) error {
	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadRequest {
		return err
	}

	period := &BudgetInvalidPeriodError{Err: err}
	if jsonErr := json.Unmarshal(apiErr.RawBody, period); jsonErr == nil && period.RequestedStartDate != "" {
		return period
	}

	// The API documents this 400 as either the flat shape above or an ordinary
	// error body with the period details on the entry, so check there too.
	for _, entry := range apiErr.Errors {
		requested, _ := entry.Extra["requested_start_date"].(string)
		if requested == "" {
			continue
		}

		previous, _ := entry.Extra["previous_valid_start_date"].(string)
		next, _ := entry.Extra["next_valid_start_date"].(string)

		return &BudgetInvalidPeriodError{
			Message:                apiErr.Message,
			ErrMsg:                 entry.Message,
			RequestedStartDate:     requested,
			PreviousValidStartDate: previous,
			NextValidStartDate:     next,
			Err:                    err,
		}
	}

	return err
}
