package http_adapter

import (
	"context"
	json2 "encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/eventstore"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/liftv2"
	"github.com/leow93/miffed-api/internal/pubsub"
)

func ensureWsConnection(t *testing.T, server *httptest.Server) *websocket.Conn {
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/socket"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {strings.TrimPrefix(server.URL, "http://")}})
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
	}
	return ws
}

func readTextMessage(t *testing.T, ws *websocket.Conn) []byte {
	mt, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("could not read message from ws connection %v", err)
	}
	if mt != websocket.TextMessage {
		t.Fatalf("expected text message, got %d", mt)
	}
	return msg
}

func TestSocket(t *testing.T) {
	t.Run("establishing a connection", func(t *testing.T) {
		store := eventstore.NewMemoryStore()
		svc := liftv2.NewLiftService(store)
		rm := liftv2.NewLiftReadModel(context.TODO(), store)

		ps := pubsub.NewMemoryPubSub()
		m := lift.NewManager(ps)
		l := lift.NewLift(ps, lift.NewLiftOpts{LowestFloor: 0, HighestFloor: 10, CurrentFloor: 0, FloorsPerSecond: 100, DoorCloseWaitMs: 1})
		m.AddLift(l)

		server := httptest.NewServer(NewServer(m, svc, rm))
		defer server.Close()
		ws := ensureWsConnection(t, server)
		defer ws.Close()

		msg := readTextMessage(t, ws)

		var event initialise
		err := json2.Unmarshal(msg, &event)
		if err != nil {
			t.Fatalf("could not unmarshal message %v", err)
		}
		if event.Type != "initialise" {
			t.Fatalf("expected initialise message, got %s", event.Type)
		}
	})

	t.Run("sending messages", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := lift.NewManager(ps)
		l := lift.NewLift(ps, lift.NewLiftOpts{LowestFloor: 0, HighestFloor: 10, CurrentFloor: 0, FloorsPerSecond: 100, DoorCloseWaitMs: 1})
		m.AddLift(l)

		store := eventstore.NewMemoryStore()
		svc := liftv2.NewLiftService(store)
		rm := liftv2.NewLiftReadModel(context.TODO(), store)
		server := httptest.NewServer(NewServer(m, svc, rm))
		defer server.Close()
		ws := ensureWsConnection(t, server)
		defer ws.Close()

		readTextMessage(t, ws) // init message

		dto := callLiftDto{
			LiftId: l.Id,
			Floor:  5,
			Type:   "call_lift",
		}
		json, _ := json2.Marshal(dto)
		if err := ws.WriteMessage(websocket.TextMessage, json); err != nil {
			t.Fatalf("could not send message over ws connection %v", err)
		}

		msg := readTextMessage(t, ws)
		var event lift.LiftMessage
		err := json2.Unmarshal(msg, &event)
		if err != nil {
			t.Fatalf("could not unmarshal message %v", err)
		}
		if event.Type != "lift_called" {
			t.Fatalf("expected lift called event, got %s", event.Type)
		}
	})
}
