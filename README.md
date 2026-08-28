# lunchmoney

[![GoDoc](https://godoc.org/github.com/icco/lunchmoney?status.svg)](https://godoc.org/github.com/icco/lunchmoney)
[![Go Report Card](https://goreportcard.com/badge/github.com/icco/lunchmoney)](https://goreportcard.com/report/github.com/icco/lunchmoney)

Go client for the [Lunch Money v2 API](https://alpha.lunchmoney.dev/introduction).

```sh
go get github.com/icco/lunchmoney
```

## Usage

Get an access token from the [developers page](https://my.lunchmoney.app/developers).

```go
client, err := lunchmoney.NewClient(os.Getenv("LUNCHMONEY_TOKEN"))
if err != nil {
	return err
}

resp, err := client.GetTransactions(ctx, &lunchmoney.TransactionFilters{
	StartDate: &start,
	EndDate:   &end,
})
if err != nil {
	return err
}

for _, t := range resp.Transactions {
	amount, err := t.ParsedAmount()
	// ...
}
```

Runnable examples are in [examples/](examples). Full API docs are on [GoDoc](https://godoc.org/github.com/icco/lunchmoney).

## Notes

Covers the stable v2 surface: users, categories, tags, transactions (with splits, groups, deletes and attachments), manual and synced crypto accounts, balance history, manual accounts, Plaid accounts, recurring items and budgets.

Amounts are strings; `ParseCurrency` turns one into a `*money.Money` scaled to the currency's minor units. Nullable IDs are pointers. Errors wrap an `*ErrorResponse` carrying the status code, reachable with `errors.As`.

v2 is in open alpha and still changing, so use a test budget while getting started.

Requires Go 1.25+. Upgrading from v0.6.x, which wrapped v1 of the API? See [UPGRADING.md](UPGRADING.md).
