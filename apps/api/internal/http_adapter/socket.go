package http_adapter

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"log"
	"net/http"
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

type callLiftDto struct {
	LiftId int32  `json:"liftId"`
	Floor  int    `json:"floor"`
	Type   string `json:"type"`
}

func reader(c *websocket.Conn, subscriptionId uuid.UUID, manager *lift.Manager) {
	defer func() {
		manager.Unsubscribe(subscriptionId)
		c.Close()
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		var req callLiftDto
		err = json.Unmarshal(message, &req)
		if err != nil {
			fmt.Println("error unmarshalling message", err)
			break
		}
		if req.Type == "call_lift" {
			manager.CallLift(req.LiftId, req.Floor)
		}
	}
}

type initialise struct {
	Type string            `json:"type"`
	Data lift.ManagerState `json:"data"`
}

func writer(c *websocket.Conn, subscriptionId uuid.UUID, ch <-chan pubsub.Message, manager *lift.Manager) {
	defer func() {
		manager.Unsubscribe(subscriptionId)
		c.Close()
	}()

	// send the current floor of the lift
	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}

	init := initialise{Type: "initialise", Data: manager.State()}
	bytes, err := json.Marshal(init)
	if err != nil {
		return
	}
	w.Write(bytes)
	if err := w.Close(); err != nil {
		return
	}

	for {
		msg := <-ch
		w, err := c.NextWriter(websocket.TextMessage)
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

		if err := w.Close(); err != nil {
			return
		}

	}
}

func socketHandler(manager *lift.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrading connection", err)
			return
		}

		id, liftChan, err := manager.Subscribe()
		if err != nil {
			log.Println("error subscribing", err)
			c.Close()
			return
		}

		go reader(c, id, manager)
		go writer(c, id, liftChan, manager)
	})
}
