package liftv3

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createLiftBody(floor int) io.Reader {
	body := createLiftReq{Floor: floor}
	bs, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	return io.Reader(bytes.NewReader(bs))
}

func Test_LiftController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc := NewLiftService(ctx)
	server := http.NewServeMux()
	server = NewController(server, svc)
	var liftId LiftId

	t.Run("POST /lift bad request results in a 400", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.Reader(strings.NewReader("{\"floor\": \"10\"}"))

		req := httptest.NewRequest("POST", "/lift", body)
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 400 {
			t.Errorf("expected 400, got %d", result.StatusCode)
		}
	})

	t.Run("POST /lift results in a 201", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/lift", createLiftBody(4))
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 201 {
			t.Errorf("expected 400, got %d", result.StatusCode)
		}
		body := createLiftRes{}
		decoder := json.NewDecoder(result.Body)
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			t.Errorf("expected no error, got %e", err)
		}
		if body.Floor != 4 {
			t.Errorf("expected 4, got %d", body.Floor)
		}
		liftId = body.Id
	})

	t.Run("GET /lift/{id}", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/lift/"+liftId.String(), io.Reader(strings.NewReader("")))
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 200 {
			t.Errorf("expected 200, got %d", result.StatusCode)
		}
		body := getLiftRes{}
		decoder := json.NewDecoder(result.Body)
		err := decoder.Decode(&body)
		if err != nil && err != io.EOF {
			t.Errorf("expected no error, got %e", err)
		}
		if body.Id != liftId {
			t.Errorf("expected %s, got %s", body.Id, liftId)
		}
	})

	t.Run("GET /lift/{id} returns 404 for unknown lift", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/lift/123", io.Reader(strings.NewReader("")))
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 404 {
			t.Errorf("expected 404 , got %d", result.StatusCode)
		}
	})

	t.Run("GET /lift returns list of lifts", func(t *testing.T) {
		// add another lift first
		liftTwo, err := svc.AddLift(LiftConfig{Floor: 5})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/lift", io.Reader(strings.NewReader("")))
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 200 {
			t.Errorf("expected 200, got %d", result.StatusCode)
		}
		body := []getLiftRes{}
		decoder := json.NewDecoder(result.Body)
		err = decoder.Decode(&body)
		if err != nil && err != io.EOF {
			t.Errorf("expected no error, got %e", err)
		}

		if len(body) != 2 {
			t.Errorf("expected 2 lifts, got %d", len(body))
		}

		if !containsId(body, liftId) {
			t.Errorf("got %s, want %s", body[0].Id, liftId)
		}

		if !containsId(body, liftTwo.Id) {
			t.Errorf("got %s, want %s", body[0].Id, liftId)
		}
	})
}

func containsId(lifts []getLiftRes, id LiftId) bool {
	for _, l := range lifts {
		if l.Id == id {
			return true
		}
	}
	return false
}
