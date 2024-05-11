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

	t.Run("LiftDoorsOpened begets LiftMessage", func(t *testing.T) {
		x := SerialiseEvent(LiftDoorsOpened{LiftId: 1, Floor: 3})
		if x == nil {
			t.Errorf("expected LiftMessage, got nil")
		}
		if x.Type != "lift_doors_opened" {
			t.Errorf("expected lift_doors_opened, got %v", x.Type)
		}
		if x.Data.(LiftDoorsOpened).LiftId != 1 {
			t.Errorf("expected 1, got %v", x.Data.(LiftDoorsOpened).LiftId)
		}
		if x.Data.(LiftDoorsOpened).Floor != 3 {
			t.Errorf("expected 3, got %v", x.Data.(LiftDoorsOpened).Floor)
		}
	})

	t.Run("LiftDoorsClosed begets LiftMessage", func(t *testing.T) {
		x := SerialiseEvent(LiftDoorsClosed{LiftId: 1, Floor: 3})
		if x == nil {
			t.Errorf("expected LiftMessage, got nil")
		}
		if x.Type != "lift_doors_closed" {
			t.Errorf("expected lift_doors_closed, got %v", x.Type)
		}
		if x.Data.(LiftDoorsClosed).LiftId != 1 {
			t.Errorf("expected 1, got %v", x.Data.(LiftDoorsClosed).LiftId)
		}
		if x.Data.(LiftDoorsClosed).Floor != 3 {
			t.Errorf("expected 3, got %v", x.Data.(LiftDoorsClosed).Floor)
		}
	})

	tests := []struct{ data any }{{data: "unknown"}, {data: map[string]string{"test": "test"}}}
	for _, test := range tests {
		t.Run("unknown type begets nil", func(t *testing.T) {
			x := SerialiseEvent(test.data)
			if x != nil {
				t.Errorf("expected nil, got %v", x)
			}
		})
	}
}
