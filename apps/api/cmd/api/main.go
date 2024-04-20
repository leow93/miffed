package main

import (
	"context"
	"fmt"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

const address = ":8080"

func main() {
	ctx := context.Background()
	ps := pubsub.NewMemoryPubSub()
	l := lift.NewLift(ctx, ps, 0, 10, 0, 1)

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		floor := l.CurrentFloor()
		w.Write([]byte(strconv.Itoa(floor)))
	})

	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		floor := rand.Intn(10)
		l.Call(floor)
		w.Write([]byte(fmt.Sprintf("floor %d called", floor)))
	})

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}
