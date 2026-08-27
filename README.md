# lunchmoney

[![GoDoc](https://godoc.org/github.com/icco/lunchmoney?status.svg)](https://godoc.org/github.com/icco/lunchmoney)
[![Go Report Card](https://goreportcard.com/badge/github.com/icco/lunchmoney)](https://goreportcard.com/report/github.com/icco/lunchmoney)

Golang API wrapper for the [Lunch Money v2 API](https://alpha.lunchmoney.dev/introduction).

To use this API, you need to create an access token on the [developers page](https://my.lunchmoney.app/developers) in the Lunch Money app.

## Notes

 - This library targets v2 of the API. v1 is no longer supported; see [Migrating](#migrating-from-v1) below.
 - The v2 API is in open alpha and still subject to change. Lunch Money suggests using a test budget while getting started.
 - We currently support read requests plus updating transactions and manual accounts. We'd love a PR to add more write support.
 - We currently only support Go 1.25 and greater.

## Migrating from v1

The v2 API is not backwards compatible with v1, so neither is this release. The changes that touch this library:

 - `GetAssets` is now `GetManualAccounts`, and `Asset` is now `ManualAccount`. `type_name` and `subtype_name` became `Type` and `Subtype`, `exclude_transactions` became `ExcludeFromTransactions`, and the `depository` type is now `cash`.
 - `GetRecurringExpenses` is now `GetRecurringItems`, and `RecurringExpense` is now `RecurringItem`. The matching criteria moved into a nested `TransactionCriteria`, overrides into `Overrides`, and match results into `Matches`. Both dates are required when filtering by range.
 - `GetBudgets` is gone. `GetBudgetSummary` reads the replacement `/summary` endpoint and returns a `BudgetSummary` rather than a list of per-category, per-month budgets. `GetBudgetSettings` is new.
 - `GetTransactions` returns a `TransactionsResponse` so that the `has_more` paging flag is visible, rather than a bare slice.
 - `GetTransaction` no longer takes filters, and `UpdateTransaction` returns the updated `Transaction` instead of an `updated` flag. Splitting moved to its own endpoint and is not wrapped here.
 - `GetCategories` takes filters. The API now defaults to the nested format, where a group carries its members in `Children`; pass `CategoryFormatFlattened` for the v1 shape.
 - On `Transaction`: `asset_id` is now `ManualAccountID`, `tags` is now `TagIDs`, `has_children` is `IsSplitParent`, `parent_id` is `SplitParentID`, `is_group` is `IsGroupParent`, and `group_id` is `GroupParentID`. Nullable IDs are pointers so that unset is distinguishable from zero.
 - v2 does not hydrate related records onto a transaction, so the category and account name fields are gone. Look them up with `GetCategory`, `GetManualAccount` or `GetPlaidAccount`.
 - Transaction status `cleared` is now `reviewed` and `uncleared` is `unreviewed`. The `pending` and `recurring` statuses are gone; use the `IsPending` field.
 - `debit_as_negative` is gone everywhere. A positive amount is always a debit and a negative amount always a credit.
 - Inserting transactions takes tag IDs, not tag names, and tags must already exist. A duplicate no longer fails the whole request: accepted rows come back in `Transactions` and rejected ones in `SkippedDuplicates`.
 - On `User`: `user_id`, `user_name` and `user_email` are now `ID`, `Name` and `Email`.
 - `ErrorResponse` matches the v2 error body, and failures now arrive with a 4xx or 5xx status instead of being buried in a 200.
 - `ParseCurrency` parses amounts as exact decimals rather than through a float, and scales to the currency's own precision. v2 returns amounts to four decimal places, which the old float path both truncated and misrounded.

## Not yet wrapped

The v2 API is much larger than v1. These endpoints exist but have no wrapper here yet: category and tag writes, transaction deletes, splitting and grouping, attachments, `POST /plaid_accounts/fetch`, budget upsert and delete, and the preview crypto and balance history endpoints.
