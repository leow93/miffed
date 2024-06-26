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
		lift.NewLift(ps, lift.NewLiftOpts{
			LowestFloor:     0,
			HighestFloor:    10,
			CurrentFloor:    0,
			FloorsPerSecond: 1,
			DoorCloseWaitMs: 500,
		}),
		lift.NewLift(ps, lift.NewLiftOpts{
			LowestFloor:     0,
			HighestFloor:    8,
			CurrentFloor:    5,
			FloorsPerSecond: 5,
			DoorCloseWaitMs: 50,
		}),
		lift.NewLift(ps, lift.NewLiftOpts{
			LowestFloor:     0,
			HighestFloor:    5,
			CurrentFloor:    3,
			FloorsPerSecond: 1,
			DoorCloseWaitMs: 2000,
		}),
		lift.NewLift(ps, lift.NewLiftOpts{
			LowestFloor:     0,
			HighestFloor:    15,
			CurrentFloor:    3,
			FloorsPerSecond: 10,
			DoorCloseWaitMs: 2000,
		}),
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
