# lunchmoney

[![GoDoc](https://godoc.org/github.com/icco/lunchmoney?status.svg)](https://godoc.org/github.com/icco/lunchmoney)
[![Go Report Card](https://goreportcard.com/badge/github.com/icco/lunchmoney)](https://goreportcard.com/report/github.com/icco/lunchmoney)

Golang API wrapper for the [Lunch Money v2 API](https://alpha.lunchmoney.dev/introduction).

Create an access token on the [developers page](https://my.lunchmoney.app/developers) to use this.

## Notes

 - Targets v2. v1 is not supported.
 - v2 is in open alpha and still changing. Use a test budget while getting started.
 - Reads, plus category writes and transaction and manual account updates. PRs welcome for the rest — see the open issues.
 - Requires Go 1.25+.

## Migrating from v1

v2 is not backwards compatible, so neither is this release.

Renames:

 - `GetAssets` → `GetManualAccounts`, `Asset` → `ManualAccount`. `type_name`/`subtype_name` → `Type`/`Subtype`, `exclude_transactions` → `ExcludeFromTransactions`, `depository` type → `cash`.
 - `GetRecurringExpenses` → `GetRecurringItems`, `RecurringExpense` → `RecurringItem`. Criteria, overrides and match results are now nested. Both dates are required when filtering by range.
 - `GetBudgets` → `GetBudgetSummary` (`/summary`), which returns a `BudgetSummary` rather than per-category, per-month budgets. `GetBudgetSettings` is new.
 - `GetCategories` takes filters. The API now defaults to nested, where a group carries its members in `Children`; pass `CategoryFormatFlattened` for the v1 shape.
 - On `Transaction`: `asset_id` → `ManualAccountID`, `tags` → `TagIDs`, `has_children` → `IsSplitParent`, `parent_id` → `SplitParentID`, `is_group` → `IsGroupParent`, `group_id` → `GroupParentID`.
 - On `User`: `user_id`/`user_name`/`user_email` → `ID`/`Name`/`Email`.

Behaviour:

 - `GetTransactions` returns a `TransactionsResponse` so the `has_more` paging flag is visible. `GetTransaction` no longer takes filters, and `UpdateTransaction` returns the updated `Transaction`. Splitting moved to its own endpoint and is not wrapped here.
 - v2 does not hydrate related records onto a transaction, so the category and account name fields are gone. Use `GetCategory`, `GetManualAccount` or `GetPlaidAccount`.
 - Status `cleared`/`uncleared` → `reviewed`/`unreviewed`. `pending` and `recurring` are gone; use `IsPending`.
 - `debit_as_negative` is gone. Positive is always a debit.
 - Inserting takes tag IDs, and tags must already exist. Duplicates no longer fail the batch: accepted rows come back in `Transactions`, rejected ones in `SkippedDuplicates`.
 - `CreateCategory` covers category groups too, behind `IsGroup`; v1's `/categories/group` is gone. `UpdateCategory`'s `Children` replaces a group's members rather than adding to them.
 - `DeleteCategory` takes a `force` argument, where v1 had a `/force` path suffix. Refusing to delete a category that is still in use returns a `CategoryDependenciesError` with the counts, which `errors.As` gets at.
 - Nullable IDs are pointers, so unset is distinguishable from zero.
 - Failures arrive with a 4xx or 5xx status instead of buried in a 200, and wrap an `ErrorResponse` — `errors.As` gets the status code and per-field errors.
 - `ParseCurrency` parses exact decimals scaled to the currency's precision. v2 returns 4 decimal places, which the old float path truncated and misrounded.
 - Amounts are sent as strings. v2 also accepts a number, but one string field avoids an empty float meaning zero.
