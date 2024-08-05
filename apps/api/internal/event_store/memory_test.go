package eventstore

import (
	"fmt"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	store := NewStore()
	msgs := []Message{
		{
			Type: "welcome",
			Data: []byte{1, 2},
		},
		{
			Type: "hello",
			Data: []byte{3, 4},
		},
	}
	t.Run("appending to a stream", func(t *testing.T) {
		result, err := store.AppendToStream("hello", msgs, 0)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
		if result.Position != 2 {
			t.Errorf("expected position of %d, got %d", 2, result.Position)
		}

		result, err = store.AppendToStream("hello", []Message{{Type: "what", Data: []byte{1, 42, 1}}}, 2)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
		if result.Position != 3 {
			t.Errorf("expected position of %d, got %d", 3, result.Position)
		}
	})

	t.Run("append to a stream concurrently", func(t *testing.T) {
		resultCh := make(chan AppendResult, 2)
		errCh := make(chan error, 2)

		for range 2 {
			fmt.Println("hello")
			go func() {
				result, err := store.AppendToStream("test'", msgs, 0)
				resultCh <- result
				errCh <- err
			}()
		}
	})
}
