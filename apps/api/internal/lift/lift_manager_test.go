package lift

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/pubsub"
)

func makeLift(ps pubsub.PubSub) *Lift {
	return NewLift(ps, NewLiftOpts{0, 10, 0, 100, 1})
}

func ensureSubscribe(t *testing.T, manager *Manager) (uuid.UUID, <-chan pubsub.Message) {
	id, ch, err := manager.Subscribe()
	if err != nil {
		t.Error("Expected to be able to subscribe")
	}
	return id, ch
}

func TestManager(t *testing.T) {
	t.Run("AddLift", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		if m.GetLift(l.Id) == nil {
			t.Error("Expected lift to be added")
		}
	})

	t.Run("Global subscription, one lift", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		id, ch := ensureSubscribe(t, m)
		defer m.Unsubscribe(id)
		l.Call(5)
		ticker := time.NewTicker(1 * time.Second)
		done := make(chan bool, 1)

		go func() {
			for {
				select {
				case <-ticker.C:
					done <- false
				case msg := <-ch:
					switch msg.(type) {
					case LiftCalled:
						done <- true
					}
				}
			}
		}()
		result := <-done
		if !result {
			t.Error("Expected to receive lift call")
		}
	})

	t.Run("Global subscription, multiple lifts", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		liftOne := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		liftTwo := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		id, ch := ensureSubscribe(t, m)
		defer m.Unsubscribe(id)
		m.CallLift(liftOne.Id, 5)
		m.CallLift(liftTwo.Id, 3)

		called := make(chan int, 2)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for {
				msg := <-ch
				switch msg.(type) {
				case LiftCalled:
					called <- msg.(LiftCalled).Floor
					wg.Done()
				}
			}
		}()

		wg.Wait()
		if len(called) != 2 {
			t.Errorf("Expected 2 lifts to be called, got %d", len(called))
		}
		close(called)
		for x := range called {
			if x != 5 && x != 3 {
				t.Errorf("Expected 5 or 3 to be called, got %d", x)
			}
		}
	})

	t.Run("adding a lift includes its events in subscriptions", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		id, ch := ensureSubscribe(t, m)
		defer func() {
			m.Unsubscribe(id)
		}()

		l := m.AddLift(NewLiftOpts{
			LowestFloor:     0,
			HighestFloor:    10,
			CurrentFloor:    5,
			FloorsPerSecond: 5,
			DoorCloseWaitMs: 1,
		})
		m.CallLift(l.Id, 4)

		called := make(chan int, 1)
		go func() {
			for {
				msg := <-ch
				fmt.Println("msg", msg)
				switch ev := msg.(type) {
				case LiftCalled:
					called <- ev.Floor
				}
			}
		}()

		select {
		case floor := <-called:
			if floor != 4 {
				t.Errorf("expected lift to be called to floor %d, got %d", 4, floor)
			}
		case <-time.After(time.Second):
			t.Error("timed out, expected to receive lift call")
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		id, ch := ensureSubscribe(t, m)
		l.Call(5)

		<-ch
		m.Unsubscribe(id)
		l.Call(3)
		select {
		case <-ch:
			t.Error("Expected to not receive lift call")
		case <-time.After(1 * time.Second):
			// ok
		}
	})

	t.Run("getting state", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l1 := m.AddLift(NewLiftOpts{0, 10, 0, 100, 1})
		l2 := m.AddLift(NewLiftOpts{10, 40, 10, 100, 1})
		state := m.State()
		if len(state) != 2 {
			t.Errorf("Expected 2 lifts, got %d", len(state))
		}
		if state[l1.Id].CurrentFloor != 0 {
			t.Errorf("Expected l1 to be on floor 0, got %d", state[l1.Id].CurrentFloor)
		}
		if state[l2.Id].CurrentFloor != 10 {
			t.Errorf("Expected l2 to be on floor 10, got %d", state[l2.Id].CurrentFloor)
		}
	})
}
