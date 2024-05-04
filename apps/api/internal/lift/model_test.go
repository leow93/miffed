package lift

import (
	"context"
	"github.com/leow93/miffed-api/internal/pubsub"
	"slices"
	"sync"
	"testing"
)

func TestCallLift(t *testing.T) {

	// Not a realistic speed, but makes testing faster
	const floorsPerSecond = 100
	const topic = "lift"

	t.Run("calling a lift", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		lift := NewLift(ps, 0, 10, floorsPerSecond)
		_, sub, _ := ps.Subscribe(topic)

		lift.Call(5)
		ev := <-sub
		liftCalled := LiftCalled{
			Floor: 5,
		}
		if ev != liftCalled {
			t.Errorf("Expected to be notified when lift is called")
		}
	})

	t.Run("calling a lift is idempotent", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		lift := NewLift(ps, 0, 10, floorsPerSecond)
		_, sub, _ := ps.Subscribe(topic)

		called := lift.Call(5)
		if !called {
			t.Errorf("Expected lift to be called")
		}
		ev := <-sub
		liftCalled := LiftCalled{
			Floor: 5,
		}
		if ev != liftCalled {
			t.Errorf("Expected to be notified when lift is called")
		}
		called = lift.Call(5)
		if called {
			t.Errorf("Expected lift to not be called")
		}
	})

	t.Run("notification of lift transits", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		lift := NewLift(ps, 0, 10, floorsPerSecond)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift.Start(ctx)
		_, sub, _ := ps.Subscribe(topic)
		lift.Call(2)
		expectedEvents := []Event{
			LiftCalled{Floor: 2},
			LiftTransited{From: 0, To: 1},
			LiftTransited{From: 1, To: 2},
			LiftArrived{Floor: 2},
		}
		for _, expected := range expectedEvents {
			ev := <-sub
			if ev != expected {
				t.Errorf("Expected %v, got %v", expected, ev)
			}
		}
		if lift.CurrentFloor() != 2 {
			t.Errorf("Expected current floor to be 2, got %d", lift.CurrentFloor())
		}
	})

	t.Run("calling a lift down", func(t *testing.T) {

		ps := pubsub.NewMemoryPubSub()
		lift := NewLift(ps, 0, 10, floorsPerSecond)
		_, sub, _ := ps.Subscribe(topic)
		lift.Call(5)
		done := make(chan bool, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift.Start(ctx)

		go func() {
			for {
				ev := <-sub
				if ev == (LiftArrived{Floor: 5}) {
					done <- true
					break
				}
			}
		}()
		<-done
		if lift.CurrentFloor() != 5 {
			t.Errorf("Expected current floor to be 5, got %d", lift.CurrentFloor())
		}
	})

	t.Run("lift visits all floors called", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		lift := NewLift(ps, 0, 10, floorsPerSecond)
		_, sub, _ := ps.Subscribe(topic)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift.Start(ctx)

		wg := sync.WaitGroup{}
		wg.Add(3)
		visited := make(chan int, 3)
		for _, floor := range []int{5, 3, 7} {
			go func() {
				lift.Call(floor)
				wg.Done()
			}()
		}
		wg.Wait()

		// Now all lifts have been called, we wait until they've been visited
		wg.Add(3)
		go func() {
			for {
				ev := <-sub
				switch ev.(type) {
				case LiftArrived:
					visited <- ev.(LiftArrived).Floor
					wg.Done()
				}
			}
		}()
		wg.Wait()
		if len(visited) != 3 {
			t.Errorf("Expected 3 floors to be visited, got %d", len(visited))
		}
		visitedSlice := make([]int, 3)
		for i := 0; i < 3; i++ {
			visitedSlice[i] = <-visited
		}
		if !slices.Contains(visitedSlice, 5) {
			t.Errorf("Expected floor 5 to be visited")
		}
		if !slices.Contains(visitedSlice, 3) {
			t.Errorf("Expected floor 3 to be visited")
		}
		if !slices.Contains(visitedSlice, 7) {
			t.Errorf("Expected floor 7 to be visited")
		}
	})
}
