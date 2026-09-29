package main

import (
	"context"
	"log"
	"os"

	"go.icco.me/lunchmoney"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("LUNCHMONEY_TOKEN")
	client, _ := lunchmoney.NewClient(token)
	ts, err := client.GetManualAccounts(ctx)
	if err != nil {
		log.Panicf("err: %+v", err)
	}

	for _, t := range ts {
		log.Printf("%+v", t) //nolint:gosec // example intentionally logs the API response
	}
}
