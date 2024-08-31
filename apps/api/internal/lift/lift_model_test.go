package lift

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/leow93/miffed-api/internal/pubsub"
)

func Test_AddingLifts(t *testing.T) {
	t.Run("it can add a lift", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)

		lift, err := svc.AddLift(LiftConfig{Floor: 10})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}
		if lift.Floor != 10 {
			t.Errorf("expected floor of 10 , got %d", lift.Floor)
		}

		getLiftRes, err := svc.GetLift(context.TODO(), lift.Id)
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}
		if getLiftRes.Id != lift.Id {
			t.Errorf("expected %s, got %s", lift.Id, getLiftRes.Id)
		}
	})

	t.Run("getting an unknown lift returns an error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)

		id := NewLiftId()

		_, err := svc.GetLift(context.TODO(), id)
		if err == nil {
			t.Error("expected an error, got nil")
			return
		}

		if !errors.Is(err, ErrLiftNotFound) {
			t.Errorf("expected lift not found error, got %e", err)
		}
	})
}

func Test_CallingLift(t *testing.T) {
	t.Run("returns an error if the lift cannot be found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)

		id := NewLiftId()

		err := svc.CallLift(context.TODO(), id, 5)
		if err == nil {
			t.Error("expected an error, got nil")
			return
		}
		if !errors.Is(err, ErrLiftNotFound) {
			t.Errorf("expected lift not found error, got %e", err)
		}
	})

	t.Run("lift can be called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)

		lift, err := svc.AddLift(LiftConfig{Floor: 10})
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}

		err = svc.CallLift(context.TODO(), lift.Id, 5)
		if err != nil {
			t.Errorf("expected no error, got %e", err)
		}
	})

	t.Run("lift descends between floors after being called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(context.TODO(), ps)
		id, ch, _ := subs.Subscribe()
		lift, err := svc.AddLift(LiftConfig{Floor: 10})
		defer func() {
			subs.Unsubscribe(id)
			cancel()
		}()
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}
		svc.CallLift(context.TODO(), lift.Id, 7)

		expectedEvents := []LiftEvent{
			{
				EventType: "lift_added",
				LiftId:    lift.Id,
				Data:      LiftAdded{Floor: 10},
			},
			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 10, To: 9},
			},
			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 9, To: 8},
			},
			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 8, To: 7},
			},
			{
				EventType: "lift_arrived",
				LiftId:    lift.Id,
				Data:      LiftArrived{Floor: 7},
			},
		}

		for _, want := range expectedEvents {
			select {
			case <-time.After(time.Second):
				t.Error("timed out")
				return
			case got := <-ch:
				if got.EventType != want.EventType {
					t.Errorf("expected %s, got %s", want.EventType, got.EventType)
					return
				}
				if got.LiftId.String() != want.LiftId.String() {
					t.Errorf("expected %s, got %s", want.LiftId.String(), got.LiftId.String())
					return
				}

				if got.Data != want.Data {
					t.Errorf("expected %T%v, got %T%v", want.Data, want.Data, got.Data, got.Data)
				}
			}
		}
	})

	t.Run("lift ascends between floors after being called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(context.TODO(), ps)
		id, ch, err := subs.Subscribe()
		lift, err := svc.AddLift(LiftConfig{Floor: 10})
		defer func() {
			subs.Unsubscribe(id)
			cancel()
		}()
		if err != nil {
			t.Errorf("expected no error, got %e", err)
			return
		}

		svc.CallLift(context.TODO(), lift.Id, 12)

		expectedEvents := []LiftEvent{
			{
				EventType: "lift_added",
				LiftId:    lift.Id,
				Data:      LiftAdded{Floor: 10},
			},

			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 10, To: 11},
			},
			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 11, To: 12},
			},
			{
				EventType: "lift_arrived",
				LiftId:    lift.Id,
				Data:      LiftArrived{Floor: 12},
			},
		}

		for _, want := range expectedEvents {
			select {
			case <-time.After(time.Second):
				t.Error("timed out")
				return
			case got := <-ch:
				if got.EventType != want.EventType {
					t.Errorf("expected %s, got %s", want.EventType, got.EventType)
					return
				}
				if got.LiftId.String() != want.LiftId.String() {
					t.Errorf("expected %s, got %s", want.LiftId.String(), got.LiftId.String())
					return
				}

				if got.Data != want.Data {
					t.Errorf("expected %T%v, got %T%v", want.Data, want.Data, got.Data, got.Data)
				}
			}
		}
	})
}

func Test_SubscriptionManager(t *testing.T) {
	t.Run("a subscription returns events", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(ctx, ps)
		id, ch, _ := subs.Subscribe()
		lift, _ := svc.AddLift(LiftConfig{Floor: 4})
		svc.CallLift(ctx, lift.Id, 5)
		defer func() {
			subs.Unsubscribe(id)
			cancel()
		}()

		expectedEvents := []LiftEvent{
			{
				EventType: "lift_added",
				LiftId:    lift.Id,
				Data:      LiftAdded{Floor: 4},
			},

			{
				EventType: "lift_transited",
				LiftId:    lift.Id,
				Data:      LiftTransited{From: 4, To: 5},
			},
			{
				EventType: "lift_arrived",
				LiftId:    lift.Id,
				Data:      LiftArrived{Floor: 5},
			},
		}

		for _, want := range expectedEvents {
			got := <-ch
			if got.EventType != want.EventType {
				t.Errorf("expected %s, got %s", want.EventType, got.EventType)
				return
			}
			if got.LiftId.String() != want.LiftId.String() {
				t.Errorf("expected %s, got %s", want.LiftId.String(), got.LiftId.String())
				return
			}

			if got.Data != want.Data {
				t.Errorf("expected %T%v, got %T%v", want.Data, want.Data, got.Data, got.Data)
			}
		}
	})

	t.Run("unsubscribing", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(ctx, ps)
		id, ch, _ := subs.Subscribe()
		lift, _ := svc.AddLift(LiftConfig{Floor: 4})
		defer func() {
			cancel()
		}()

		msg := <-ch
		if msg.EventType != "lift_added" {
			t.Errorf("expected lift_added, got %s", msg.EventType)
			return
		}
		subs.Unsubscribe(id)
		svc.CallLift(context.TODO(), lift.Id, 10)

		select {
		case ev := <-ch:
			t.Errorf("expected no message, got %v", ev)
		case <-time.After(100 * time.Millisecond):
			// ok
		}
	})

	t.Run("events are sent in order to the subscriber", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ps := pubsub.NewMemoryPubSub()
		svc := NewLiftService(ctx, ps)
		subs := NewSubscriptionManager(ctx, ps)
		lift, _ := svc.AddLift(LiftConfig{Floor: 0})
		id, ch, _ := subs.Subscribe()
		defer func() {
			subs.Unsubscribe(id)
			cancel()
		}()

		// pull off the lift_added event
		ev := <-ch
		if ev.EventType != "lift_added" {
			t.Errorf("expected lift_added, got %s", ev.EventType)
		}

		go func() {
			for i := 0; i < 50; i++ {
				err := svc.CallLift(context.TODO(), lift.Id, i+1)
				if err != nil {
					fmt.Println("err", err)
				}
			}
		}()

		var want []LiftEvent
		for i := 0; i < 50; i++ {
			want = append(want, createLiftEvent(lift.Id, "lift_transited", LiftTransited{From: i, To: i + 1}))
			want = append(want, createLiftEvent(lift.Id, "lift_arrived", LiftArrived{Floor: i + 1}))
		}

		wg := sync.WaitGroup{}
		wg.Add(100)
		var got []LiftEvent
		go func() {
			for i := 0; i < 100; i++ {
				ev := <-ch
				got = append(got, ev)
				wg.Done()
			}
		}()

		wg.Wait()
		if len(got) != len(want) {
			t.Errorf("expected %d events, got %d", len(want), len(got))
		}
		for i := 0; i < len(want); i++ {
			if got[i].EventType != want[i].EventType {
				t.Errorf("expected %s, got %s", got[i].EventType, want[i].EventType)
				return
			}

			if got[i].Data != want[i].Data {
				t.Errorf("expected %T%v, got %T%v", got[i].Data, got[i].Data, want[i].Data, want[i].Data)
				return
			}
		}
	})
}
