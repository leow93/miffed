package liftv2

import (
	"context"
	"fmt"
	"testing"

	"github.com/leow93/miffed-api/internal/eventstore"
)

func given(evs []LiftEvent, cmd command) []LiftEvent {
	state := LiftModel{}
	for _, ev := range evs {
		state = evolve(state, ev)
	}

	newEvs := decide(cmd, state)

	return newEvs
}

func TestLift_decider(t *testing.T) {
	t.Run("adding a lift", func(t *testing.T) {
		newEvs := given([]LiftEvent{}, AddLift{Id: NewLiftId(), Floor: 1})
		if len(newEvs) != 1 {
			t.Fatalf("expected 1 event, got %d", len(newEvs))
		}

		ev := newEvs[0]
		if ev.eventType() != "lift_added" {
			t.Fatalf("expected lift_added, got %s", ev.eventType())
		}
	})

	t.Run("adding a lift is idempotent", func(t *testing.T) {
		newEvs := given([]LiftEvent{LiftAdded{Id: NewLiftId(), Floor: 1}}, AddLift{Id: NewLiftId(), Floor: 1})
		if len(newEvs) != 0 {
			t.Fatalf("expected 0 event, got %d", len(newEvs))
		}
	})
}

func TestLift_integration(t *testing.T) {
	store := eventstore.NewMemoryStore()
	svc := NewLiftService(store)
	t.Run("adding a lift writes to the correct stream", func(t *testing.T) {
		ctx := context.TODO()
		id := NewLiftId()
		svc.AddLift(ctx, AddLift{Id: id, Floor: 12})

		streamEvs, _, err := store.ReadStream(fmt.Sprintf("Lift-%s", id))
		if err != nil {
			t.Fatalf("expected no error, got %e", err)
		}
		if len(streamEvs) != 1 {
			t.Errorf("expected one event, got %d", len(streamEvs))
		}
	})

	t.Run("we only add the lift once", func(t *testing.T) {
		ctx := context.TODO()
		id := NewLiftId()
		svc.AddLift(ctx, AddLift{Id: id, Floor: 12})
		svc.AddLift(ctx, AddLift{Id: id, Floor: 12})
		streamEvs, _, err := store.ReadStream(streamName(id))
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}
		if len(streamEvs) != 1 {
			t.Errorf("expected one event, got %d", len(streamEvs))
		}
	})
}
