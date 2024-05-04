package lift

import "testing"

func TestSerialiseEvent(t *testing.T) {
	t.Run("nil begets nil", func(t *testing.T) {
		x := SerialiseEvent(nil)
		if x != nil {
			t.Errorf("expected nil, got %v", x)
		}
	})

	t.Run("LiftArrived begets LiftMessage", func(t *testing.T) {
		x := SerialiseEvent(LiftArrived{Floor: 1})
		if x == nil {
			t.Errorf("expected LiftMessage, got nil")
		}
		if x.Type != "lift_arrived" {
			t.Errorf("expected lift_arrived, got %v", x.Type)
		}
		if x.Data.(LiftArrived).Floor != 1 {
			t.Errorf("expected 1, got %v", x.Data.(LiftArrived).Floor)
		}
	})

	t.Run("LiftCalled begets LiftMessage", func(t *testing.T) {
		x := SerialiseEvent(LiftCalled{Floor: 1})
		if x == nil {
			t.Errorf("expected LiftMessage, got nil")
		}
		if x.Type != "lift_called" {
			t.Errorf("expected lift_called, got %v", x.Type)
		}
		if x.Data.(LiftCalled).Floor != 1 {
			t.Errorf("expected 1, got %v", x.Data.(LiftCalled).Floor)
		}
	})

	t.Run("LiftTransited begets LiftMessage", func(t *testing.T) {
		x := SerialiseEvent(LiftTransited{From: 1, To: 2})
		if x == nil {
			t.Errorf("expected LiftMessage, got nil")
		}
		if x.Type != "lift_transited" {
			t.Errorf("expected lift_transited, got %v", x.Type)
		}
		if x.Data.(LiftTransited).From != 1 {
			t.Errorf("expected 1, got %v", x.Data.(LiftTransited).From)
		}
		if x.Data.(LiftTransited).To != 2 {
			t.Errorf("expected 2, got %v", x.Data.(LiftTransited).To)
		}
	})

	t.Run("unknown type begets nil", func(t *testing.T) {
		x := SerialiseEvent("unknown")
		if x != nil {
			t.Errorf("expected nil, got %v", x)
		}
	})
}
