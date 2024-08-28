package liftv3

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

func writer(c *websocket.Conn, manager *SubscriptionManager, sub *subscription) {
	defer func() {
		manager.Unsubscribe(sub.Id)
		c.Close()
	}()

	for {
		msg := <-sub.EventsCh
		w, err := c.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		fmt.Println("msg", msg)

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

func socketHandler(subscriptionMgr *SubscriptionManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrading connection", err)
			return
		}

		sub := subscriptionMgr.Subscribe()

		// go reader(c, svc)
		go writer(c, subscriptionMgr, sub)
	})
}

func NewSocket(mux *http.ServeMux, subs *SubscriptionManager) *http.ServeMux {
	mux.Handle("/socket", socketHandler(subs))
	return mux
}
