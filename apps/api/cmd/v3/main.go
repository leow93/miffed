package main

import (
	"context"
	"log"
	"net/http"

	"github.com/leow93/miffed-api/internal/liftv3"
	"github.com/leow93/miffed-api/internal/pubsub"
)

func callLift(svc *liftv3.LiftService, id liftv3.LiftId, floor int) {
	ctx := context.TODO()
	err := svc.CallLift(ctx, id, floor)
	if err != nil {
		panic(err)
	}
}

const address = ":8080"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ps := pubsub.NewMemoryPubSub()
	svc := liftv3.NewLiftService(ctx, ps)
	subs := liftv3.NewSubscriptionManager(ctx, ps)

	mux := http.NewServeMux()
	mux = liftv3.NewController(mux, svc)
	mux = liftv3.NewSocket(mux, subs)

	if err := http.ListenAndServe(address, mux); err != nil {
		cancel()
		log.Fatal(err)
	}
}
