package http_adapter

import (
	json2 "encoding/json"
	"github.com/gorilla/websocket"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		ps := pubsub.NewMemoryPubSub()
		l := lift.NewLift(ps, 0, 10, 0, 1)

		server := httptest.NewServer(NewServer(l, ps))
		defer server.Close()
		ws := ensureWsConnection(t, server)
		defer ws.Close()

		msg := readTextMessage(t, ws)

		var event currentFloor
		err := json2.Unmarshal(msg, &event)
		if err != nil {
			t.Fatalf("could not unmarshal message %v", err)
		}
		if event.Type != "current_floor" {
			t.Fatalf("expected current floor message, got %s", event.Type)
		}
	})

	t.Run("sending messages", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		l := lift.NewLift(ps, 0, 10, 0, 1)

		server := httptest.NewServer(NewServer(l, ps))
		defer server.Close()
		ws := ensureWsConnection(t, server)
		defer ws.Close()

		readTextMessage(t, ws) // current floor message

		dto := newCallLiftDto(5)
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
