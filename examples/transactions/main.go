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
	client, _ := lunchmoney.NewClient(token)
	resp, err := client.GetTransactions(ctx, nil)
	if err != nil {
		log.Panicf("err: %+v", err)
	}

	for _, t := range resp.Transactions {
		log.Printf("%+v", t) //nolint:gosec // example intentionally logs the API response
	}

	if resp.HasMore {
		log.Printf("more transactions available; request the next page with an offset")
	}
}
