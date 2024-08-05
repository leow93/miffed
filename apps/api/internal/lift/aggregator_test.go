package lift

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leow93/miffed-api/internal/pubsub"
)

func incrementlog(name string) func() {
	var count Id = 0
	return func() {
		atomic.AddInt32(&count, 1)
		fmt.Printf("\n%s: %d\n", name, count)
	}
}

func TestAggregator(t *testing.T) {
	testLiftOpts := NewLiftOpts{
		LowestFloor:     0,
		HighestFloor:    100,
		CurrentFloor:    0,
		FloorsPerSecond: 1000,
		DoorCloseWaitMs: 0,
	}
	t.Run("lift added events are sent to the sink", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		repo := NewLiftRepo(ps)
		sinkChan := make(chan LiftMessage)
		errChan := make(chan error)
		NewAggregator(ctx, ps, repo, sinkChan, errChan)
		repo.AddLift(testLiftOpts)
		select {
		case <-time.After(5 * time.Second):
			t.Error("timed out")
		case msg := <-sinkChan:
			if msg.Type != "lift_added" {
				t.Errorf("expected lift_added, got %s", msg.Type)
			}
		case e := <-errChan:
			t.Errorf("something went wrong: %s", e.Error())
		}
	})

	t.Run("lift deleted events are sent to the sink", func(t *testing.T) {
		ps := pubsub.NewMemoryPubSub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		repo := NewLiftRepo(ps)
		lift := repo.AddLift(testLiftOpts)
		sinkChan := make(chan LiftMessage)
		errChan := make(chan error)
		NewAggregator(ctx, ps, repo, sinkChan, errChan)
		repo.DeleteLift(lift.Id)
		select {
		case <-time.After(time.Second):
			t.Error("timed out")
		case msg := <-sinkChan:
			if msg.Type != "lift_deleted" {
				t.Errorf("expected lift_deleted, got %s", msg.Type)
			}
		case e := <-errChan:
			t.Errorf("something went wrong: %s", e.Error())
		}
	})

	t.Run("lift motion events are sent to the sink", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		repo := NewLiftRepo(ps)
		lift := repo.AddLift(testLiftOpts)
		lift.Start(ctx)
		sinkChan := make(chan LiftMessage)
		errChan := make(chan error)
		NewAggregator(ctx, ps, repo, sinkChan, errChan)
		go func() {
			e := <-errChan
			panic(e)
		}()

		var evs []LiftMessage
		wg := sync.WaitGroup{}
		wg.Add(6)
		go func() {
			for {
				msg := <-sinkChan
				evs = append(evs, msg)
				wg.Done()
			}
		}()
		lift.Call(2)
		wg.Wait()
		cancel()

		if len(evs) != 6 {
			t.Errorf("expected %d events, got %d", 6, len(evs))
		}

		// just check for arrival, don't want to test the lift implementation details too much
		arrived := false
		for _, ev := range evs {
			switch ev.Type {
			case "lift_arrived":
				arrived = true
			}
		}
		if !arrived {
			t.Error("expected lift to arrive at some point")
		}
	})

	// TODO: test what happens when ctx is cancelled
}
