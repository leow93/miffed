package liftv3

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
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

func Test_Socket(t *testing.T) {
	t.Run("establishing a connection", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(ctx, ps)

		mux := http.NewServeMux()
		mux = NewSocket(mux, subs)
		server := httptest.NewServer(mux)
		ws := ensureWsConnection(t, server)
		defer ws.Close()

		svc.AddLift(LiftConfig{Floor: 5})

		msg := readTextMessage(t, ws)
		var event LiftEvent
		err := json.Unmarshal(msg, &event)
		if err != nil {
			t.Fatalf("could not unmarshal message %v", err)
		}
		if event.EventType != "lift_added" {
			t.Errorf("expected lift_added, got %s", event.EventType)
		}
	})
}
