package liftv2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/eventstore"
)

type LiftSpeed struct {
	FloorsPerSecond int `json:"floors_per_second"`
}

type lifecycle int

const (
	initial lifecycle = iota
	created
)

type LiftId struct {
	uuid.UUID
}

func NewLiftId() LiftId {
	return LiftId{uuid.New()}
}

func ParseLiftId(id string) (LiftId, error) {
	uuid, err := uuid.Parse(id)
	return LiftId{uuid}, err
}

// State
type LiftModel struct {
	Id        LiftId
	Floor     int
	Speed     LiftSpeed
	lifecycle lifecycle
}

// Commands

// All commands must implement this interface
type command interface {
	commandType() string
	id() string
}
type AddLift struct {
	Floor int
	Id    LiftId
}

func (AddLift) commandType() string {
	return "add_lift"
}

func (l AddLift) id() string {
	return streamName(l.Id)
}

// Domain Events, must implement the interface
type LiftEvent interface {
	eventType() string
	serialise() ([]byte, error)
}

type LiftAdded struct {
	Id    LiftId    `json:"id"`
	Floor int       `json:"floor"`
	Speed LiftSpeed `json:"speed"`
}

func (LiftAdded) eventType() string {
	return "lift_added"
}

func (ev LiftAdded) serialise() ([]byte, error) {
	return json.Marshal(ev)
}

type Event struct {
	eventType string
	data      []byte
}

type LiftRepo interface {
	AddLift(ctx context.Context, lift AddLift) (LiftModel, error)
}

type eventStore interface {
	AppendToStream(streamName string, expectedVersion uint64, events []eventstore.Event) error
	ReadStream(streamName string) ([]eventstore.Event, uint64, error)
}

type LiftService struct {
	decider eventstore.DecisionFunc[command, LiftEvent]
}

func NewLiftService(store eventStore) *LiftService {
	return &LiftService{
		decider: LiftDecider(store),
	}
}

func streamName(liftId LiftId) string {
	return fmt.Sprintf("Lift-%s", liftId.String())
}

func (svc *LiftService) AddLift(ctx context.Context, cmd AddLift) error {
	return svc.decider(cmd)
}

func evolve(state LiftModel, event LiftEvent) LiftModel {
	if state.lifecycle == initial && event.eventType() == "lift_added" {
		ev := event.(LiftAdded)
		return LiftModel{
			lifecycle: created,
			Id:        ev.Id,
			Floor:     ev.Floor,
			Speed:     ev.Speed,
		}
	}
	return state
}

func decide(command command, state LiftModel) []LiftEvent {
	var evs []LiftEvent

	if state.lifecycle == initial && command.commandType() == "add_lift" {
		addLift := command.(AddLift)
		evs = append(evs, LiftAdded{Id: addLift.Id, Floor: addLift.Floor, Speed: LiftSpeed{FloorsPerSecond: 100}})
	}

	return evs
}

func streamId(cmd command) string {
	return cmd.id()
}

func deserialise(ev eventstore.Event) (LiftEvent, error) {
	switch ev.EventType {
	case "lift_added":
		var result LiftAdded
		err := json.Unmarshal(ev.Data, &result)
		if err != nil {
			return nil, err
		}

		return result, nil
	default:
		{
			return nil, fmt.Errorf("unknown event type: %s", ev.EventType)
		}
	}
}

func serialise(ev LiftEvent) (eventstore.Event, error) {
	var resultEv eventstore.Event
	bytes, err := ev.serialise()
	if err != nil {
		return resultEv, err
	}

	resultEv = eventstore.Event{
		EventType: ev.eventType(),
		Data:      bytes,
	}

	return resultEv, nil
}

func LiftDecider(
	store eventStore,
) eventstore.DecisionFunc[command, LiftEvent] {
	initialState := LiftModel{
		lifecycle: initial,
	}
	return eventstore.NewDecider(store, initialState, evolve, decide, streamId, deserialise, serialise)
}
