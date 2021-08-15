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

	ts, err := client.GetBudgets(ctx)
	if err != nil {
		log.Fatalf("get err: %+v", err)
	}

	for _, t := range ts {
		log.Printf("%+v", t)
	}
}
