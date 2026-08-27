package main

import (
	"context"
	"log"
	"os"

	"github.com/icco/lunchmoney"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("LUNCHMONEY_TOKEN")
	client, err := lunchmoney.NewClient(token)
	if err != nil {
		log.Fatalf("client err: %+v", err)
	}

	includeTotals := true
	opts := &lunchmoney.BudgetFilters{
		StartDate:     "2021-01-01",
		EndDate:       "2021-12-31",
		IncludeTotals: &includeTotals,
	}

	summary, err := client.GetBudgetSummary(ctx, opts)
	if err != nil {
		log.Fatalf("get err: %+v", err)
	}

	log.Printf("aligned: %t totals: %+v", summary.Aligned, summary.Totals) //nolint:gosec // example intentionally logs the API response
	for _, c := range summary.Categories {
		log.Printf("%+v", c) //nolint:gosec // example intentionally logs the API response
	}
}
