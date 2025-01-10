package main

import (
	"context"
	"log"
	"os"

	"github.com/rshep3087/lunchmoney"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("LUNCHMONEY_TOKEN")
	client, err := lunchmoney.NewClient(token)
	if err != nil {
		log.Fatalf("client err: %+v", err)
	}

	opts := &lunchmoney.BudgetFilters{
		StartDate: "2021-01-01",
		EndDate:   "2021-12-31",
	}

	ts, err := client.GetBudgets(ctx, opts)
	if err != nil {
		log.Fatalf("get err: %+v", err)
	}

	for _, t := range ts {
		log.Printf("%+v", t)
	}
}
