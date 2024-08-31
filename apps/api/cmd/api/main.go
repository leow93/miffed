package main

import (
	"context"
	"log"
	"net/http"

	"github.com/leow93/miffed-api/internal/httpadapter"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"github.com/rs/cors"
)

func callLift(svc *lift.LiftService, id lift.LiftId, floor int) {
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
	svc := lift.NewLiftService(ctx, ps)
	subs := lift.NewSubscriptionManager(ctx, ps)

	mux := http.NewServeMux()
	mux = httpadapter.NewController(mux, svc)
	mux = httpadapter.NewSocket(mux, subs)

	server := cors.AllowAll().Handler(mux)

	if err := http.ListenAndServe(address, server); err != nil {
		cancel()
		log.Fatal(err)
	}
}
