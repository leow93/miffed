package lift

import (
	"context"
	"slices"
	"sync"
	"testing"
)

func TestCallLift(t *testing.T) {

	// Not a realistic speed, but makes testing faster
	const floorsPerSecond = 100

	t.Run("calling a lift", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift := NewLift(ctx, 0, 10, 0, floorsPerSecond)
		here := lift.Call(5)
		<-here
		if lift.currentFloor != 5 {
			t.Errorf("Expected current floor to be 5, got %d", lift.currentFloor)
		}
	})

	t.Run("calling a lift down", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift := NewLift(ctx, 0, 10, 10, floorsPerSecond)
		here := lift.Call(5)
		<-here
		if lift.currentFloor != 5 {
			t.Errorf("Expected current floor to be 5, got %d", lift.currentFloor)
		}
	})

	t.Run("lift visits all floors called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		lift := NewLift(ctx, 0, 10, 0, floorsPerSecond)
		wg := sync.WaitGroup{}
		wg.Add(3)
		visited := make(chan int, 3)

		for _, floor := range []int{5, 3, 7} {
			go func() {
				here := lift.Call(floor)
				<-here
				visited <- floor
				wg.Done()
			}()
		}
		wg.Wait()
		if len(visited) != 3 {
			t.Errorf("Expected 3 floors to be visited, got %d", len(visited))
		}
		close(visited)
		var floors []int
		for f := range visited {
			floors = append(floors, f)
		}
		if !slices.Contains(floors, 5) {
			t.Errorf("Expected floor 5 to be visited")
		}
		if !slices.Contains(floors, 3) {
			t.Errorf("Expected floor 3 to be visited")
		}
		if !slices.Contains(floors, 7) {
			t.Errorf("Expected floor 7 to be visited")
		}
	})

	t.Skip("lift calling is idempotent")

	t.Skip("concurrent calls")
}
