package liftv3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/leow93/miffed-api/internal/pubsub"
)

func waitFor[T any](f func() (T, error), timer <-chan time.Time) (T, error) {
	for {
		select {
		case <-timer:
			var x T
			return x, fmt.Errorf("timed out")
		default:
			result, err := f()
			if err != nil {
				return waitFor(f, timer)
			}
			return result, nil

		}
	}
}

func waitForLiftAtFloor(svc *LiftService, liftId LiftId, floor int) (Lift, error) {
	fn := func() (Lift, error) {
		l, err := svc.GetLift(context.TODO(), liftId)
		if err != nil {
			return Lift{}, err
		}
		if l.Floor != floor {
			return Lift{}, errors.New("wrong floor")
		}
		return l, nil
	}
	timer := time.After(time.Second)
	return waitFor(fn, timer)
}

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
	ps := pubsub.NewMemoryPubSub()
	svc := NewLiftService(ctx, ps)
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

	t.Run("GET /lift returns a consistent order", func(t *testing.T) {
		var lifts []Lift
		for i := 0; i < 100; i++ {
			lift, err := svc.AddLift(LiftConfig{Floor: 5})
			if err != nil {
				lifts = append(lifts, lift)
			}
		}

		expectedOrder := []LiftId{}
		// test order ten times
		for i := 0; i < 10; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/lift", io.Reader(strings.NewReader("")))
			server.ServeHTTP(rec, req)
			result := rec.Result()
			if result.StatusCode != 200 {
				t.Errorf("expected 200, got %d", result.StatusCode)
			}
			body := []getLiftRes{}
			decoder := json.NewDecoder(result.Body)
			err := decoder.Decode(&body)
			if err != nil && err != io.EOF {
				t.Errorf("expected no error, got %e", err)
			}

			// Fill with result of first pass
			if len(expectedOrder) == 0 {
				for _, l := range body {
					expectedOrder = append(expectedOrder, l.Id)
				}
			} else {
				// test against expected order
				if len(expectedOrder) != len(body) {
					t.Errorf("expected %d, got %d", len(expectedOrder), len(body))
					return
				}
				for i := 0; i < len(expectedOrder); i++ {
					if expectedOrder[i] != body[i].Id {
						t.Errorf("expected %s, got %s", expectedOrder[i], body[i].Id)
						return
					}
				}
			}
		}
	})

	t.Run("POST /lift/{id}/call calls the lift to the given floor", func(t *testing.T) {
		rec := httptest.NewRecorder()
		body := io.Reader(strings.NewReader("{\"floor\": 100}"))
		req := httptest.NewRequest("POST", "/lift/"+liftId.String()+"/call", body)
		server.ServeHTTP(rec, req)

		result := rec.Result()
		if result.StatusCode != 201 {
			t.Errorf("expected 201, got %d", result.StatusCode)
		}

		lift, err := waitForLiftAtFloor(svc, liftId, 100)
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}
		if lift.Floor != 100 {
			t.Errorf("expected %d, got %d", 100, lift.Floor)
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
