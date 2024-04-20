package lift

import (
	"testing"
)

func TestCallLift(t *testing.T) {

	// Not a realistic speed, but makes testing faster
	const floorsPerSecond = 100

	t.Run("calling a lift", func(t *testing.T) {
		lift := NewLift(0, 10, 0, floorsPerSecond)
		here := lift.Call(5)
		<-here
		if lift.currentFloor != 5 {
			t.Errorf("Expected current floor to be 5, got %d", lift.currentFloor)
		}
	})
}
