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

	categories, err := client.GetCategories(ctx)
	if err != nil {
		log.Fatalf("get err: %+v", err)
	}

	for _, category := range categories {
		log.Printf("%+v\n", category) //nolint:gosec // example intentionally logs the API response
	}

	log.Printf("Fetching category details: %s\n", categories[0].Name) //nolint:gosec // example intentionally logs the API response

	category, err := client.GetCategory(ctx, categories[0].ID)
	if err != nil {
		log.Fatalf("get err: %+v", err)
	}
	log.Printf("%+v\n", category) //nolint:gosec // example intentionally logs the API response
}
