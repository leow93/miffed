package http_adapter

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"log"
	"net/http"
)

type Client struct {
	conn *websocket.Conn
	lift *lift.Lift
	ps   pubsub.PubSub
}

func (client *Client) sendLiftUpdates() {
	id, liftChan, err := client.ps.Subscribe("lift")
	defer func() {
		client.conn.Close()
		client.ps.Unsubscribe(id)
	}()

	if err != nil {
		client.conn.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	for {
		msg := <-liftChan
		w, err := client.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		content := lift.SerialiseEvent(msg)

		if content != nil {
			bytes, err := json.Marshal(*content)
			if err != nil {
				return
			}
			w.Write(bytes)
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if r.Host == "localhost:8080" {
			return true
		}

		originHeader := r.Header.Get("Origin")
		return originHeader == r.Host
	},
}

// TODO: handle client closing connection
func socketHandler(lift *lift.Lift, ps pubsub.PubSub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Upgrading connection")
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrading connection", err)
			return
		}

		client := Client{c, lift, ps}

		go client.sendLiftUpdates()
	})
}
