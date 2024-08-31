package httpadapter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/lift"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if r.Host == "localhost:8080" {
			return true
		}

		originHeader := r.Header.Get("Origin")
		return originHeader == r.Host
	},
}

func writer(c *websocket.Conn, manager *lift.SubscriptionManager, id uuid.UUID, ch <-chan lift.LiftEvent) {
	defer func() {
		manager.Unsubscribe(id)
		c.Close()
	}()

	for {
		msg := <-ch
		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		bytes, err := json.Marshal(&msg)
		if err != nil {
			return
		}
		w.Write(bytes)

		if err := w.Close(); err != nil {
			return
		}

	}
}

func socketHandler(subscriptionMgr *lift.SubscriptionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrading connection", err)
			return
		}

		id, ch, err := subscriptionMgr.Subscribe()
		if err != nil {
			log.Println("error subscribing", err)
			return
		}

		go writer(c, subscriptionMgr, id, ch)
	})
}

func NewSocket(mux *http.ServeMux, subs *lift.SubscriptionManager) *http.ServeMux {
	mux.Handle("/socket", socketHandler(subs))
	return mux
}
