# lunchmoney

[![GoDoc](https://godoc.org/github.com/icco/lunchmoney?status.svg)](https://godoc.org/github.com/icco/lunchmoney)
[![Go Report Card](https://goreportcard.com/badge/github.com/icco/lunchmoney)](https://goreportcard.com/report/github.com/icco/lunchmoney)

Go wrapper for the [Lunch Money v2 API](https://alpha.lunchmoney.dev/introduction). Requires Go 1.25+.

Get an access token from the [developers page](https://my.lunchmoney.app/developers).

```go
client, err := lunchmoney.NewClient(os.Getenv("LUNCHMONEY_TOKEN"))
if err != nil {
	return err
}

txns, err := client.GetTransactions(ctx, &lunchmoney.TransactionFilters{
	StartDate: &start,
	EndDate:   &end,
})
```

See [examples/](examples) for more.

## Coverage

Covers the stable v2 surface: users, categories, tags, transactions (including splits, groups, deletes and attachments), manual accounts, Plaid accounts, recurring items and budgets.

The preview crypto and balance history endpoints are not wrapped — see [#34](https://github.com/icco/lunchmoney/issues/34).

v1 is not supported. v2 is in open alpha and still changing, so use a test budget while getting started.

## Conventions

 - Amounts are strings, parsed with `ParseCurrency` into exact decimals scaled to the currency's precision. v2 returns four decimal places.
 - Nullable IDs are pointers, so unset is distinguishable from zero.
 - Date-only fields are strings; timestamps are `time.Time`.
 - Failures carry a 4xx or 5xx status and wrap an `ErrorResponse`. `errors.As` gets the status code and per-field errors. Some endpoints wrap something richer — `CategoryDependenciesError`, `TagInUseError`, `BudgetInvalidPeriodError` — and `IsTooEarly` reports a Plaid fetch already in flight.

## Migrating from v1

v2 is not backwards compatible, so neither is this library from v0.6.x.

Renamed:

 - `GetAssets` → `GetManualAccounts`, `Asset` → `ManualAccount`. `type_name`/`subtype_name` → `Type`/`Subtype`, `exclude_transactions` → `ExcludeFromTransactions`, and the `depository` type is now `cash`.
 - `GetRecurringExpenses` → `GetRecurringItems`, `RecurringExpense` → `RecurringItem`, with criteria, overrides and matches now nested.
 - `GetBudgets` → `GetBudgetSummary`, returning a `BudgetSummary` rather than per-category, per-month budgets.
 - On `Transaction`: `asset_id` → `ManualAccountID`, `tags` → `TagIDs`, `has_children` → `IsSplitParent`, `parent_id` → `SplitParentID`, `is_group` → `IsGroupParent`, `group_id` → `GroupParentID`.
 - On `User`: `user_id`/`user_name`/`user_email` → `ID`/`Name`/`Email`.

Changed behaviour:

 - `debit_as_negative` is gone. A positive amount is always a debit.
 - Status `cleared`/`uncleared` → `reviewed`/`unreviewed`. `pending` and `recurring` are gone; use `IsPending`.
 - Transactions no longer carry hydrated category and account names. Look them up with `GetCategory`, `GetManualAccount` or `GetPlaidAccount`.
 - `GetTransactions` returns a `TransactionsResponse`, so the `has_more` paging flag is visible. `GetTransaction` takes no filters, and `UpdateTransaction` returns the updated transaction.
 - `InsertTransactions` takes tag IDs, and the tags must already exist. Duplicates no longer fail the batch: accepted rows come back in `Transactions`, rejected ones in `SkippedDuplicates`.
 - `GetCategories` takes filters and defaults to the nested format, where a group carries its members in `Children`. Pass `CategoryFormatFlattened` for the v1 shape.
 - `CreateCategory` covers groups behind `IsGroup`; v1's `/categories/group` is gone. `UpdateCategory`'s `Children` replaces a group's members rather than adding to them.
 - `DeleteCategory` and `DeleteTag` take a `force` argument where v1 used a `/force` path suffix.
 - Errors arrive with a real status code instead of buried in a 200.
