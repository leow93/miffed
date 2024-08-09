package queue

import "testing"

func TestQueue(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(5)
		if q.Length() != 1 {
			t.Fatalf("expected queue length to be 1, got %d", q.Length())
		}
	})
	t.Run("dequeue", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(5)
		floor, err := q.Dequeue()
		if err != nil {
			t.Fatalf("expected value to be 5, got error %e", err)
		}
		if floor != 5 {
			t.Fatalf("expected value to be 5, got %d", floor)
		}
		if q.Length() != 0 {
			t.Fatalf("expected queue length to be 0, got %d", q.Length())
		}
	})

	t.Run("queue is ordered", func(t *testing.T) {
		q := NewQueue()
		q.Enqueue(5)
		q.Enqueue(3)
		q.Enqueue(4)

		expected := []int{5, 3, 4}

		for _, want := range expected {
			got, err := q.Dequeue()
			if err != nil {
				t.Fatalf("expected value to be %d, got error %e", want, err)
			}
			if got != want {
				t.Fatalf("expected value to be %d, got %d", want, got)
			}
		}
	})

	t.Run("dequeue empty queue returns an error", func(t *testing.T) {
		q := NewQueue()
		_, err := q.Dequeue()
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("queue can be checked for a value", func(t *testing.T) {
		q := NewQueue()
		if q.Has(1) == true {
			t.Error("expected false, got true")
		}
		q.Enqueue(1)
		if q.Has(1) == false {
			t.Error("expected true, got false")
		}
	})
}
