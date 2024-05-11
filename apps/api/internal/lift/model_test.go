package lift

import (
	"context"
	"errors"
	"github.com/leow93/miffed-api/internal/pubsub"
	"slices"
	"sync"
	"testing"
	"time"
)

// Not a realistic floorsPerSecond, but makes testing faster
const floorsPerSecond = 100

func createLift() *Lift {
	ps := pubsub.NewMemoryPubSub()
	return NewLift(ps, 0, 10, floorsPerSecond)
}

func subscribe(t *testing.T, lift *Lift) <-chan pubsub.Message {
	_, ch, err := lift.Subscribe()
	if err != nil {
		t.Errorf("Error subscribing to lift %e", err)
	}
	return ch
}

func TestCallLift(t *testing.T) {

	t.Run("calling a lift", func(t *testing.T) {
		lift := createLift()
		sub := subscribe(t, lift)

		lift.Call(5)
		ev := <-sub
		liftCalled := LiftCalled{
			LiftId: lift.Id,
			Floor:  5,
		}
		if ev != liftCalled {
			t.Errorf("Expected to be notified when lift is called")
		}
	})

	t.Run("calling a lift is idempotent", func(t *testing.T) {
		lift := createLift()
		sub := subscribe(t, lift)
		called := lift.Call(5)
		if !called {
			t.Errorf("Expected lift to be called")
		}
		ev := <-sub
		liftCalled := LiftCalled{
			LiftId: lift.Id,
			Floor:  5,
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
		lift := createLift()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift.Start(ctx)
		sub := subscribe(t, lift)
		lift.Call(2)
		expectedEvents := []Event{
			LiftCalled{LiftId: lift.Id, Floor: 2},
			LiftTransited{LiftId: lift.Id, From: 0, To: 1},
			LiftTransited{LiftId: lift.Id, From: 1, To: 2},
			LiftArrived{LiftId: lift.Id, Floor: 2},
		}
		for _, expected := range expectedEvents {
			ev := <-sub

			if ev != expected {
				t.Errorf("Expected %v, got %v", expected, ev)
			}
		}
		if lift.State().CurrentFloor != 2 {
			t.Errorf("Expected current floor to be 2, got %d", lift.State().CurrentFloor)
		}
	})

	t.Run("calling a lift down", func(t *testing.T) {
		lift := createLift()
		sub := subscribe(t, lift)
		lift.Call(5)
		done := make(chan bool, 1)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift.Start(ctx)

		go func() {
			for {
				ev := <-sub
				if ev == (LiftArrived{LiftId: lift.Id, Floor: 5}) {
					done <- true
					break
				}
			}
		}()
		<-done
		if lift.State().CurrentFloor != 5 {
			t.Errorf("Expected current floor to be 5, got %d", lift.State().CurrentFloor)
		}
	})

	t.Run("lift visits all floors called", func(t *testing.T) {
		lift := createLift()
		sub := subscribe(t, lift)
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

func TestDoors(t *testing.T) {
	t.Run("doors open on arrival", func(t *testing.T) {
		lift := createLift()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sub := subscribe(t, lift)
		lift.Start(ctx)
		lift.Call(1)

		timer := time.After(time.Second)
		errCh := make(chan error, 1)
		msgCh := make(chan LiftDoorsOpened, 1)

		go func() {
			for {
				select {
				case <-timer:
					errCh <- errors.New("Message not received in time")
				case msg := <-sub:
					switch msg.(type) {
					case LiftDoorsOpened:
						msgCh <- msg.(LiftDoorsOpened)
					}
				}
			}

		}()

		for range 1 {
			select {
			case e := <-errCh:
				t.Errorf("Expected message, received error: \"%s\"", e)
			case <-msgCh:
				// ok
			}
		}
	})

	t.Run("doors close after opening", func(t *testing.T) {
		lift := createLift()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sub := subscribe(t, lift)
		lift.Start(ctx)
		lift.Call(1)

		ticker := time.NewTicker(time.Millisecond * 50)
		errCh := make(chan error, 1)
		msgCh := make(chan pubsub.Message, 2)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for {
				select {
				case <-ticker.C:
					errCh <- errors.New("message not received in time")
					wg.Done()
				case msg := <-sub:
					switch msg.(type) {
					case LiftDoorsOpened:
						msgCh <- msg
						wg.Done()
					case LiftDoorsClosed:
						msgCh <- msg
						wg.Done()
					}
				}
			}
		}()

		wg.Wait()
		<-msgCh
		closed := <-msgCh
		switch closed.(type) {
		case LiftDoorsClosed:
		// ok
		default:
			t.Errorf("Received wrong message")
		}

	})
}

func TestNewId(t *testing.T) {
	t.Run("generating an id is thread safe", func(t *testing.T) {
		current := NewId()
		wg := sync.WaitGroup{}
		wg.Add(1000)
		for i := 0; i < 1000; i++ {
			go func() {
				NewId()
				wg.Done()
			}()
		}
		wg.Wait()
		id := NewId()
		expected := current + 1001
		if id != expected {
			t.Errorf("Expected %d to be generated, got %d", expected, id)
		}
	})
}
