package http_adapter

import (
	"encoding/json"
	"fmt"
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
	Floor int    `json:"floor"`
	Type  string `json:"type"`
}

func newCallLiftDto(floor int) callLiftDto {
	return callLiftDto{Floor: floor, Type: "call_lift"}
}

func reader(c *websocket.Conn, lift *lift.Lift) {
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
			called := lift.Call(req.Floor)
			if called {
				fmt.Println("lift called")
			}
		}
	}
}

type currentFloor struct {
	Floor int    `json:"floor"`
	Type  string `json:"type"`
}

func writer(c *websocket.Conn, l *lift.Lift, ps pubsub.PubSub) {
	id, liftChan, err := ps.Subscribe("lift")
	if err != nil {
		return
	}
	defer ps.Unsubscribe(id)

	// send the current floor of the lift
	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}
	var cf currentFloor
	cf.Floor = l.CurrentFloor()
	cf.Type = "current_floor"
	bytes, err := json.Marshal(cf)
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

func socketHandler(lift *lift.Lift, ps pubsub.PubSub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("error upgrading connection", err)
			return
		}

		go reader(c, lift)
		go writer(c, lift, ps)
	})
}
