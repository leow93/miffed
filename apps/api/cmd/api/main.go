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
	liftManager := lift.NewManager(ps)

	lifts := []*lift.Lift{
		lift.NewLift(ps, 0, 10, 1),
		lift.NewLift(ps, 0, 8, 1),
		lift.NewLift(ps, 0, 15, 3),
		lift.NewLift(ps, 0, 20, 2),
		lift.NewLift(ps, 0, 40, 4),
	}

	for _, l := range lifts {
		liftManager.AddLift(l)
		l.Start(ctx)
	}

	server := http_adapter.NewServer(liftManager)

	if err := http.ListenAndServe(address, server); err != nil {
		log.Fatal(err)
	}
}
