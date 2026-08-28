# Upgrading from v0.6.x to v0.7.x

v0.6.x wrapped v1 of the Lunch Money API. v0.7.x wraps v2, which is not backwards compatible, so neither is this library. v1 is no longer supported.

Lunch Money's own [migration guide](https://alpha.lunchmoney.dev/v2/migration-guide) covers the API side. This covers the Go side.

## Client

`BaseAPIURL` is now `https://api.lunchmoney.dev/v2/`.

Failures arrive with a real 4xx or 5xx status instead of a 200 with an error in the body, and wrap an `*ErrorResponse` carrying the status code and per-field errors:

```go
var apiErr *lunchmoney.ErrorResponse
if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
	// ...
}
```

Some calls wrap something more specific: `*CategoryDependenciesError`, `*TagInUseError` and `*BudgetInvalidPeriodError`. `IsTooEarly(err)` reports a Plaid fetch already in flight.

`ParseCurrency` now parses exact decimals scaled to the currency's minor units. v2 returns amounts to four decimal places, which the old `int64(100 * float)` path truncated and misrounded, so parsed values may differ by a cent from v0.6.x.

## Renamed

| v0.6.x | v0.7.x |
| --- | --- |
| `GetAssets`, `UpdateAsset` | `GetManualAccounts`, `GetManualAccount`, `UpdateManualAccount` |
| `Asset` | `ManualAccount` |
| `GetRecurringExpenses` | `GetRecurringItems` |
| `RecurringExpense` | `RecurringItem` |
| `GetBudgets` | `GetBudgetSummary` |

On `ManualAccount`: `TypeName`/`SubtypeName` are `Type`/`Subtype`, `ExcludedTransactions` is `ExcludeFromTransactions`, and the `depository` type value is now `cash`.

On `Transaction`: `AssetID` is `ManualAccountID`, `Tags` is `TagIDs` (IDs, not objects), `HasChildren` is `IsSplitParent`, `ParentID` is `SplitParentID`, `IsGroup` is `IsGroupParent`, `GroupID` is `GroupParentID`.

On `User`: `UserID`, `UserName` and `UserEmail` are `ID`, `Name` and `Email`.

## Signatures

`GetCategories` takes filters, and defaults to the nested format where a group carries its members in `Children`. Pass `CategoryFormatFlattened` for the v1 shape:

```go
cats, err := client.GetCategories(ctx, &lunchmoney.CategoryFilters{
	Format: lunchmoney.CategoryFormatFlattened,
})
```

`GetTransactions` returns a `*TransactionsResponse` rather than a slice, so the `has_more` paging flag is visible:

```go
resp, err := client.GetTransactions(ctx, filters)
for _, t := range resp.Transactions {
	// ...
}
```

`GetTransaction` no longer takes filters. `UpdateTransaction` returns the updated `*Transaction` instead of an `updated` flag. `GetRecurringItems` requires `StartDate` and `EndDate` together, where v1 accepted a start date alone. `GetBudgetSummary` requires both dates.

## Behaviour

`debit_as_negative` is gone everywhere, from filters and request bodies alike. A positive amount is always a debit and a negative one always a credit, regardless of account preference.

Transaction status `cleared` and `uncleared` are now `reviewed` and `unreviewed`, as the `StatusReviewed` and `StatusUnreviewed` constants. The `pending` and `recurring` statuses are gone; pending transactions are flagged by `IsPending`.

Transactions no longer arrive with related records hydrated onto them. `CategoryName`, `CategoryGroupID`, `CategoryGroupName`, `IsIncome`, `ExcludeFromBudget`, `ExcludeFromTotals`, `AssetName`, `PlaidAccountName`, `InstitutionName` and the rest are gone — resolve the IDs with `GetCategory`, `GetManualAccount` or `GetPlaidAccount`.

`InsertTransactions` takes tag IDs, and the tags must already exist; v1 created them inline from names. Duplicates no longer fail the whole batch: accepted rows come back in `Transactions`, rejected ones in `SkippedDuplicates` with a reason.

`RecurringItem` nests what v1 kept flat. `StartDate`, `EndDate`, `Payee`, `Amount` and the cadence live under `TransactionCriteria`; the values applied to matches live under `Overrides`; match results live under `Matches`, which is nil for suggested items. `BillingDate` is `TransactionCriteria.AnchorDate`.

`GetBudgetSummary` returns a `BudgetSummary` for a date range rather than v1's per-category, per-month `Budget` list. `GetBudgetSettings` reports the account's budget period configuration.

Nullable IDs are pointers, so unset is distinguishable from zero. This affects `Transaction.CategoryID`, `RecurringID`, `ManualAccountID`, `PlaidAccountID`, `SplitParentID` and `GroupParentID`, and `Category.GroupID`.

Splitting moved out of `UpdateTransaction` to `SplitTransaction`, and unsplitting to `UnsplitTransaction`. Grouping is `GroupTransactions` and `UngroupTransactions`.

`CreateCategory` covers category groups behind `IsGroup`; v1's separate group endpoint is gone. `UpdateCategory`'s `Children` replaces a group's members rather than adding to them. `DeleteCategory` and `DeleteTag` take a `force` argument where v1 used a `/force` path suffix.

## New in v0.7.x

Category, tag and manual account writes; transaction deletes, splits, groups and attachments; budget upsert and delete; `TriggerPlaidFetch`; and single-item getters for tags, plaid accounts, manual accounts and recurring items.

Crypto and balance history endpoints are now wrapped, including manual/synced crypto and per-account/per-symbol balance history variants.
