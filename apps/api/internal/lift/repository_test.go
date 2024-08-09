package lift

import (
	"sync"
	"testing"

	"github.com/leow93/miffed-api/internal/pubsub"
)

func TestLiftRepository(t *testing.T) {
	ps := pubsub.NewMemoryPubSub()
	t.Run("adding a lift", func(t *testing.T) {
		repo := NewLiftRepo(ps)
		lift := repo.AddLift(NewLiftOpts{})
		l := repo.GetLift(lift.Id)
		if l == nil {
			t.Error("expected lift, got nil")
		}

		lifts := repo.GetLifts()
		if len(lifts) != 1 {
			t.Errorf("expected one lift, got %d", len(lifts))
		}
	})

	t.Run("adding a lift is thread-safe", func(t *testing.T) {
		repo := NewLiftRepo(ps)
		wg := sync.WaitGroup{}
		wg.Add(1000)

		for range 1000 {
			go func() {
				repo.AddLift(NewLiftOpts{})
				wg.Done()
			}()
		}
		wg.Wait()
		lifts := repo.GetLifts()
		if len(lifts) != 1000 {
			t.Errorf("expected 1000 lifts, got %d", len(lifts))
		}
	})

	t.Run("deleting a lift", func(t *testing.T) {
		repo := NewLiftRepo(ps)
		lift1 := repo.AddLift(NewLiftOpts{})
		lift2 := repo.AddLift(NewLiftOpts{})
		count := len(repo.GetLifts())
		if count != 2 {
			t.Errorf("expected two lifts, got %d", len(repo.GetLifts()))
		}
		repo.DeleteLift(lift2.Id)
		lifts := repo.GetLifts()
		count = len(lifts)
		if count != 1 {
			t.Errorf("expected one lift, got %d", count)
		}
		id := lifts[0].Id
		if id != lift1.Id {
			t.Errorf("expected %d, got %d", lift1.Id, id)
		}
	})
}

func getSubscription(ps pubsub.PubSub) (<-chan pubsub.Message, func()) {
	id, ch, err := ps.Subscribe(LiftRepoTopic)
	if err != nil {
		panic(err)
	}

	return ch, func() {
		ps.Unsubscribe(id)
	}
}

func TestLiftRepositoryPublishing(t *testing.T) {
	ps := pubsub.NewMemoryPubSub()

	t.Run("adding a lift publishes a message", func(t *testing.T) {
		repo := NewLiftRepo(ps)
		ch, unsubscribe := getSubscription(ps)
		defer unsubscribe()
		lift := repo.AddLift(NewLiftOpts{})
		msg := <-ch
		liftAdded := msg.(LiftAdded)
		if liftAdded.LiftId != lift.Id {
			t.Errorf("expected %d, got %d", lift.Id, liftAdded.LiftId)
		}
	})

	t.Run("deleting a lift publishes a message", func(t *testing.T) {
		repo := NewLiftRepo(ps)
		lift := repo.AddLift(NewLiftOpts{})
		ch, unsubscribe := getSubscription(ps)
		defer unsubscribe()
		repo.DeleteLift(lift.Id)
		msg := <-ch
		liftDeleted := msg.(LiftDeleted)
		if liftDeleted.LiftId != lift.Id {
			t.Errorf("expected %d, got %d", lift.Id, liftDeleted.LiftId)
		}
	})
}
