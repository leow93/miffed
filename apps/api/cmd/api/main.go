package main

import (
	"context"
	"github.com/leow93/miffed-api/internal/http_adapter"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"log"
	"net/http"
)

const address = ":8080"

func main() {
	ctx := context.Background()
	ps := pubsub.NewMemoryPubSub()
	l := lift.NewLift(ps, 0, 10, 1)
	l.Start(ctx)

	server := http_adapter.NewServer(l, ps)

	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatal(err)
	}
}
