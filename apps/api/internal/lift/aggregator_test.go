package lift

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/leow93/miffed-api/internal/pubsub"
)

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

	t.Run("adding a new lift to the aggregator", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		repo := NewLiftRepo(ps)
		sinkChan := make(chan LiftMessage)
		errChan := make(chan error)
		NewAggregator(ctx, ps, repo, sinkChan, errChan)
		lift := repo.AddLift(testLiftOpts)
		lift.Start(ctx)
		select {
		case msg := <-sinkChan:
			if msg.Type != "lift_added" {
				t.Errorf("expected %s, got %s", "lift_added", msg.Type)
			}
		case <-time.After(time.Second):
			t.Error("timed out")
		case e := <-errChan:
			t.Errorf("something went wrong: %e", e)
		}

		var ev LiftMessage
		lift.Call(1)
		ev = <-sinkChan

		if ev.Type != "lift_called" {
			t.Errorf("expected %s, got %s", "lift_called", ev.Type)
		}
	})

	t.Run("cancelling context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		repo := NewLiftRepo(ps)
		sinkChan := make(chan LiftMessage)
		errChan := make(chan error)
		NewAggregator(ctx, ps, repo, sinkChan, errChan)
		lift := repo.AddLift(testLiftOpts)
		lift.Start(ctx)
		select {
		case msg := <-sinkChan:
			if msg.Type != "lift_added" {
				t.Errorf("expected %s, got %s", "lift_added", msg.Type)
			}
		case <-time.After(time.Second):
			t.Error("timed out")
		case e := <-errChan:
			t.Errorf("something went wrong: %e", e)
		}

		go func() {
			for i := range 10 {
				lift.Call(i)
			}
			cancel()
		}()

		var liftCalledEvs []LiftMessage
		wg := sync.WaitGroup{}
		wg.Add(10)
		go func() {
			for {
				select {
				case ev := <-sinkChan:
					if ev.Type == "lift_called" {
						liftCalledEvs = append(liftCalledEvs, ev)
						wg.Done()
					}
				case err := <-errChan:
					t.Errorf("something went wrong: %e", err)
				}
			}
		}()
		wg.Wait()
		if len(liftCalledEvs) != 10 {
			t.Errorf("expected %d events, got %d", 10, len(liftCalledEvs))
		}
	})
}
