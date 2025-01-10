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
	client, _ := lunchmoney.NewClient(token)
	ts, err := client.GetPlaidAccounts(ctx)
	if err != nil {
		log.Panicf("err: %+v", err)
	}

	for _, t := range ts {
		log.Printf("%+v", t)
	}
}
