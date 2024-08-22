package liftv2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/leow93/miffed-api/internal/eventstore"
)

func Test_AddLift(t *testing.T) {
	ctx := context.Background()
	store := eventstore.NewMemoryStore()
	svc := NewLiftService(store)
	readModel := NewLiftReadModel(ctx, store)
	server := http.NewServeMux()
	server = NewController(server, svc, readModel)
	var liftId LiftId

	t.Run("it can create a lift", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.Reader(strings.NewReader("{}"))

		req := httptest.NewRequest("POST", "/lift", body)
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 201 {
			t.Errorf("expected 201, got %d", result.StatusCode)
		}

		var res LiftModel
		json.NewDecoder(result.Body).Decode(&res)
		liftId = res.Id
	})

	t.Run("it can get a lift", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.Reader(strings.NewReader(""))
		req := httptest.NewRequest("GET", "/lift/"+liftId.String(), body)
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 200 {
			t.Errorf("expected 200, got %d", result.StatusCode)
		}
		var res LiftModel
		json.NewDecoder(result.Body).Decode(&res)
		if res.Id != liftId {
			t.Errorf("expected lift id of %s, got %s", liftId, res.Id)
		}
	})

	t.Run("it can get all lifts", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.Reader(strings.NewReader(""))
		// Read model is eventually consistent so let's wait for a bit...
		time.Sleep(100 * time.Millisecond)
		req := httptest.NewRequest("GET", "/lift", body)
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 200 {
			t.Errorf("expected 200, got %d", result.StatusCode)
		}
		var res []LiftModel
		json.NewDecoder(result.Body).Decode(&res)
		if len(res) != 1 {
			t.Errorf("expected 1 lift, got %d", len(res))
		}
	})
}
