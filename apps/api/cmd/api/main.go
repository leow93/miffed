package main

import (
	"context"
	"fmt"
	"github.com/leow93/miffed-api/internal/lift"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

const address = ":8080"

type CallRequest struct {
	Floor int `json:"floor"`
}

func main() {
	ctx := context.Background()
	l := lift.NewLift(ctx, 0, 10, 0, 1)

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
