package liftv3

import (
	"github.com/google/uuid"
)

type LiftId struct{ uuid.UUID }

func NewLiftId() LiftId {
	return LiftId{uuid.New()}
}

func ParseLiftId(id string) (LiftId, error) {
	ID, err := uuid.Parse(id)
	if err != nil {
		return LiftId{}, nil
	}

	return LiftId{ID}, nil
}

type LiftEvent struct {
	Data      any
	EventType string
	LiftId    LiftId
}

func createLiftEvent(liftId LiftId, eventType string, data any) LiftEvent {
	return LiftEvent{
		Data:      data,
		EventType: eventType,
		LiftId:    liftId,
	}
}

type LiftAdded struct {
	Floor int
}

type LiftCalled struct {
	From int
	To   int
}

type LiftTransited struct {
	From int
	To   int
}

type LiftArrived struct {
	Floor int
}
