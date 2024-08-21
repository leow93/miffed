package liftv2

import (
	"context"
	"encoding/json"
	"fmt"
)

type LiftSpeed struct {
	FloorsPerSecond int `json:"floors_per_second"`
}

type LiftModel struct {
	Id    int
	Floor int
	Speed LiftSpeed
}

type LiftAdded struct {
	Id    int       `json:"id"`
	Floor int       `json:"floor"`
	Speed LiftSpeed `json:"speed"`
}

func (ev LiftAdded) serialise() ([]byte, error) {
	return json.Marshal(ev)
}

type LiftEvent[T any] struct {
	eventType string
	data      T
}

type Event struct {
	eventType string
	data      []byte
}

func serialise[T any](ev LiftEvent[T]) (*Event, error) {
	switch ev.eventType {
	case "lift_added":
		data, err := json.Marshal(ev.data)
		return &Event{eventType: ev.eventType, data: data}, err
	default:
		return nil, fmt.Errorf("unknown event type: %s", ev.eventType)
	}
}

type LiftRepo interface {
	AddLift(ctx context.Context, lift NewLiftOpts) (LiftModel, error)
}

type eventStore interface {
	StreamVersion(streamName string) (uint64, error)
	AppendToStream(streamName string, expectedVersion uint64, events []Event) error
}

type LiftService struct {
	repo  LiftRepo
	store eventStore
}

type NewLiftOpts struct {
	floor int
}

func streamName(liftId int) string {
	return fmt.Sprintf("Lift-%d", liftId)
}

func (svc *LiftService) publish(ctx context.Context, liftId int, ev Event) error {
	stream := streamName(liftId)
	version, err := svc.store.StreamVersion(stream)
	if err != nil {
		return err
	}

	return svc.store.AppendToStream(stream, version, []Event{event})
}

func (svc *LiftService) AddLift(ctx context.Context, opts NewLiftOpts) error {
	_, err := svc.repo.AddLift(ctx, opts)
	if err != nil {
		return err
	}
	return nil
}
