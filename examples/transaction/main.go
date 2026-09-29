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
	t, err := client.GetTransaction(ctx, 1)
	if err != nil {
		log.Panicf("err: %+v", err)
	}

	log.Printf("%+v", t) //nolint:gosec // example intentionally logs the API response
}
