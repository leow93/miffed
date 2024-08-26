package liftv3

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_AddingLifts(t *testing.T) {
	t.Run("it can add a lift", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		svc := NewLiftService(ctx)

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
		svc := NewLiftService(ctx)

		id := NewLiftId()

		_, err := svc.GetLift(context.TODO(), id)
		if err == nil {
			t.Error("expected an error, got nil")
			return
		}

		got := err.Error()
		want := fmt.Sprintf("lift not found: %s", id.String())

		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	})
}

func Test_CallingLift(t *testing.T) {
	t.Run("returns an error if the lift cannot be found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		svc := NewLiftService(ctx)

		id := NewLiftId()

		err := svc.CallLift(context.TODO(), id, 5)
		if err == nil {
			t.Error("expected an error, got nil")
			return
		}

		got := err.Error()
		want := fmt.Sprintf("lift not found: %s", id.String())

		if got != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	})

	t.Run("lift can be called", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		svc := NewLiftService(ctx)

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
		defer cancel()
		svc := NewLiftService(ctx)

		lift, err := svc.AddLift(LiftConfig{Floor: 10})
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
			case got := <-svc.Notifications:
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
		defer cancel()
		svc := NewLiftService(ctx)

		lift, err := svc.AddLift(LiftConfig{Floor: 10})
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
			case got := <-svc.Notifications:
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
