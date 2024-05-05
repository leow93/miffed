package http_adapter

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/lift"
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
	LiftId string `json:"liftId"`
	Floor  int    `json:"floor"`
	Type   string `json:"type"`
}

func reader(c *websocket.Conn, manager *lift.Manager) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		var req callLiftDto
		err = json.Unmarshal(message, &req)
		if err != nil {
			break
		}
		if req.Type == "call_lift" {
			id, e := lift.ParseId(req.LiftId)
			if e != nil {
				break
			}
			manager.CallLift(id, req.Floor)
		}
	}
}

type initialise struct {
	Type string            `json:"type"`
	Data lift.ManagerState `json:"data"`
}

func writer(c *websocket.Conn, manager *lift.Manager) {
	id, liftChan, err := manager.Subscribe()
	if err != nil {
		return
	}
	defer manager.Unsubscribe(id)

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
		msg := <-liftChan
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

		go reader(c, manager)
		go writer(c, manager)
	})
}
