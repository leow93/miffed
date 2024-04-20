package main

import (
	"log"
	"net/http"
)

const address = ":8080"

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("miffed"))
	})

	if err := http.ListenAndServe(address, handler); err != nil {
		log.Fatal(err)
	}
}
