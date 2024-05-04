package manager

import (
	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"sync"
	"testing"
	"time"
)

func makeLift(ps pubsub.PubSub) *lift.Lift {
	return lift.NewLift(ps, 0, 10, 1)
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
		l := makeLift(ps)
		m.AddLift(l)
		if m.GetLift(l.Id) == nil {
			t.Error("Expected lift to be added")
		}
	})

	t.Run("Global subscription, one lift", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l := makeLift(ps)
		m.AddLift(l)
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
					case lift.LiftCalled:
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
		liftOne := makeLift(ps)
		liftTwo := makeLift(ps)
		m.AddLift(liftOne)
		m.AddLift(liftTwo)
		id, ch := ensureSubscribe(t, m)
		defer m.Unsubscribe(id)
		m.CallLift(liftOne.Id, 5)
		m.CallLift(liftTwo.Id, 3)

		called := make(chan int, 2)
		ticker := time.NewTicker(1 * time.Second)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			for {
				select {
				case <-ticker.C:
					t.Errorf("Expected to receive lift call")
				case msg := <-ch:
					switch msg.(type) {
					case lift.LiftCalled:
						called <- msg.(lift.LiftCalled).Floor
						wg.Done()
					}
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

	t.Run("unsubscribe", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		m := NewManager(ps)
		l := makeLift(ps)
		m.AddLift(l)
		id, ch := ensureSubscribe(t, m)
		l.Call(5)

		<-ch
		m.Unsubscribe(id)
		l.Call(3)
		select {
		case msg := <-ch:
			if msg != nil {
				t.Error("Expected to not receive lift call")
			}
		case <-time.After(1 * time.Second):
			t.Error("Expected to not receive lift call")
		}

	})
}
