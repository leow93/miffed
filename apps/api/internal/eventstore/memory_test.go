package eventstore

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func readStreamUnsafe(store *MemoryStore, stream string) []Event {
	evs, _, err := store.ReadStream(stream)
	if err != nil {
		panic(err)
	}
	return evs
}

func TestMemoryStore(t *testing.T) {
	t.Run("stream must have a parseable category", func(t *testing.T) {
		store := NewMemoryStore()
		err := store.AppendToStream("test", 0, []Event{{eventType: "created", data: []byte("123")}})
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})

	t.Run("appending one event to a stream", func(t *testing.T) {
		store := NewMemoryStore()
		err := store.AppendToStream("test-1", 0, []Event{{eventType: "created", data: []byte("123")}})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}

		evs := readStreamUnsafe(store, "test-1")
		if len(evs) != 1 {
			t.Fatalf("expected 1 event, got %d", len(evs))
		}
	})

	t.Run("appending one event to a stream with wrong version", func(t *testing.T) {
		store := NewMemoryStore()
		err := store.AppendToStream("test-1", 1, []Event{{eventType: "created", data: []byte("123")}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		msg := err.Error()
		want := "wrong expected version: got 0, want 1"
		if msg != want {
			t.Errorf("wrong error. got %s, want %s", msg, want)
		}
	})

	t.Run("appending multiple events", func(t *testing.T) {
		store := NewMemoryStore()
		events := []Event{
			{
				eventType: "created",
				data:      []byte("123"),
			},
			{
				eventType: "updated",
				data:      []byte("456"),
			},
		}
		err := store.AppendToStream("test-1", 0, events)
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}

		evs := readStreamUnsafe(store, "test-1")
		if len(evs) != 2 {
			t.Fatalf("expected 2 events, got %d", len(evs))
		}

		events = []Event{
			{
				eventType: "deleted",
				data:      []byte(""),
			},
		}

		err = store.AppendToStream("test-1", 2, events)
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}

		evs = readStreamUnsafe(store, "test-1")
		if len(evs) != 3 {
			t.Fatalf("expected 3 events, got %d", len(evs))
		}
	})

	t.Run("concurrent writes to the same stream cause an error", func(t *testing.T) {
		store := NewMemoryStore()
		event := Event{eventType: "created", data: []byte("")}

		wg := sync.WaitGroup{}
		wg.Add(5)
		errChan := make(chan error, 4)

		for range [5]int{} {
			go func() {
				err := store.AppendToStream("test-1", 0, []Event{event})
				if err != nil {
					errChan <- err
				}
				wg.Done()
			}()
		}

		wg.Wait()

		if len(errChan) != 4 {
			t.Errorf("expected 4 errors, got %d", len(errChan))
		}

		for range 4 {
			e := <-errChan
			msg := e.Error()

			if !strings.HasPrefix(msg, "wrong expected version:") {
				t.Fatalf("got the wrong error: %s", msg)
			}
		}
	})

	t.Run("reading the category", func(t *testing.T) {
		store := NewMemoryStore()
		err := store.AppendToStream("test-1", 0, []Event{{eventType: "created", data: []byte("123")}})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}
		err = store.AppendToStream("test-2", 0, []Event{{eventType: "created", data: []byte("234")}})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}

		evs, err := store.ReadCategory("test", 0)
		if err != nil {
			t.Fatalf("expected no error, got %e", err)
		}
		if len(evs) != 2 {
			t.Fatalf("expected 2 events, got %d", len(evs))
		}
	})

	t.Run("reading nil category gives an error", func(t *testing.T) {
		store := NewMemoryStore()
		_, err := store.ReadCategory("test", 0)
		msg := err.Error()
		if msg != "no such category" {
			t.Fatalf("got the wrong error: %s", msg)
		}
	})

	t.Run("reading the category from a position", func(t *testing.T) {
		store := NewMemoryStore()
		for i := range [50]int{} {
			stream := fmt.Sprintf("test-%d", i)
			store.AppendToStream(
				stream,
				0,
				[]Event{{eventType: "test", data: []byte(strconv.Itoa(i))}})
		}

		evs, _ := store.ReadCategory("test", 40)
		if len(evs) != 10 {
			t.Fatalf("expected %d events, got %d", 10, len(evs))
		}
	})
}

func TestParseCategory(t *testing.T) {
	t.Run("fails when there is no category", func(t *testing.T) {
		_, err := parseCategory("foo")
		if err.Error() != "no category" {
			t.Fatal("expected a category")
		}
	})

	t.Run("fails when the category is empty", func(t *testing.T) {
		_, err := parseCategory("-123")
		if err.Error() != "empty category" {
			t.Fatal("expected a non-empty category")
		}
	})
}
