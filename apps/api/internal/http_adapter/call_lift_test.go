package http_adapter

import (
	"encoding/json"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func createBody[T any](body T) io.Reader {
	b, _ := json.Marshal(body)
	return io.Reader(strings.NewReader(string(b)))
}

func TestCallLift(t *testing.T) {
	t.Run("calling a lift", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		l := lift.NewLift(ps, 0, 10, 0, 1)
		server := NewServer(l, ps)

		rec := httptest.NewRecorder()
		b := createBody(callLiftReq{Floor: 5})
		req := httptest.NewRequest("POST", "/call", b)

		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 201 {
			t.Errorf("expected status 201, got %d", rec.Code)
		}
		var res callLiftRes
		json.NewDecoder(result.Body).Decode(&res)
		if res.Floor != 5 {
			t.Errorf("expected floor 5, got %d", res.Floor)
		}
	})

	t.Run("bad request begets a bad request response", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		l := lift.NewLift(ps, 0, 10, 0, 1)
		server := NewServer(l, ps)

		rec := httptest.NewRecorder()
		b := createBody(map[string]interface{}{"floor": "5"})
		req := httptest.NewRequest("POST", "/call", b)

		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 400 {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
		var res ErrResponse
		json.NewDecoder(result.Body).Decode(&res)
		if res.Code != 400 {
			t.Errorf("expected code 400, got %d", res.Code)
		}
	})
}
