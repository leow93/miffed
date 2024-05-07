package http_adapter

import (
	"encoding/json"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func createBody[T any](body T) io.Reader {
	b, _ := json.Marshal(body)
	return io.Reader(strings.NewReader(string(b)))
}

func initServer() (http.Handler, *lift.Lift) {
	ps := pubsub.NewMemoryPubSub()
	m := lift.NewManager(ps)
	l := lift.NewLift(ps, 0, 10, 1)
	m.AddLift(l)
	server := NewServer(m)
	return server, l
}

func path(liftId lift.Id) string {
	return "/lift/" + strconv.Itoa(int(liftId)) + "/call"
}

func TestCallLift(t *testing.T) {
	t.Run("calling a lift", func(t *testing.T) {
		server, l := initServer()

		rec := httptest.NewRecorder()
		b := createBody(callLiftReq{Floor: 5})

		req := httptest.NewRequest("POST", path(l.Id), b)

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
		server, l := initServer()

		rec := httptest.NewRecorder()
		b := createBody(map[string]interface{}{"floor": "5"})
		req := httptest.NewRequest("POST", path(l.Id), b)

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
