package liftv2

import (
	"testing"
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
		newEvs := given([]LiftEvent{}, AddLift{Id: 1, Floor: 1})
		if len(newEvs) != 1 {
			t.Fatalf("expected 1 event, got %d", len(newEvs))
		}

		ev := newEvs[0]
		if ev.eventType() != "lift_added" {
			t.Fatalf("expected lift_added, got %s", ev.eventType())
		}
	})

	t.Run("adding a lift is idempotent", func(t *testing.T) {
		newEvs := given([]LiftEvent{LiftAdded{Id: 1, Floor: 1}}, AddLift{Id: 1, Floor: 1})
		if len(newEvs) != 0 {
			t.Fatalf("expected 0 event, got %d", len(newEvs))
		}
	})
}
