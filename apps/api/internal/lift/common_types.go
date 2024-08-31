package lift

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
	Data      any    `json:"data"`
	EventType string `json:"event_type"`
	LiftId    LiftId `json:"lift_id"`
}

func createLiftEvent(liftId LiftId, eventType string, data any) LiftEvent {
	return LiftEvent{
		Data:      data,
		EventType: eventType,
		LiftId:    liftId,
	}
}

type LiftAdded struct {
	Floor int `json:"floor"`
}

type LiftTransited struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type LiftArrived struct {
	Floor int `json:"floor"`
}
