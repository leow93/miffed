package lift

import (
	"context"
	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/pubsub"
	"strconv"
	"sync/atomic"
	"time"
)

type Id = int32

var count Id = 0

func NewId() Id {
	return atomic.AddInt32(&count, 1)
}

func ParseId(s string) (Id, error) {
	x, err := strconv.Atoi(s)
	return int32(x), err
}

type Lift struct {
	Id              Id
	lowestFloor     int
	highestFloor    int
	currentFloor    int
	floorsPerSecond int // queue per second
	doorCloseWaitMs int
	requests        *Queue // queue to visit
	pubsub          pubsub.PubSub
}

type LiftState struct {
	Id           Id  `json:"id"`
	CurrentFloor int `json:"currentFloor"`
	LowestFloor  int `json:"lowestFloor"`
	HighestFloor int `json:"highestFloor"`
}

func topic(liftId Id) pubsub.Topic {
	return pubsub.Topic("lift:" + strconv.Itoa(int(liftId)))
}

type NewLiftOpts struct {
	LowestFloor     int
	HighestFloor    int
	CurrentFloor    int
	FloorsPerSecond int
	DoorCloseWaitMs int
}

func NewLift(ps pubsub.PubSub, opts NewLiftOpts) *Lift {
	lift := &Lift{
		Id:              NewId(),
		lowestFloor:     opts.LowestFloor,
		highestFloor:    opts.HighestFloor,
		currentFloor:    opts.CurrentFloor,
		floorsPerSecond: opts.FloorsPerSecond,
		doorCloseWaitMs: opts.DoorCloseWaitMs,
		requests:        NewQueue(),
		pubsub:          ps,
	}
	return lift
}

func (l *Lift) enqueue(floor int) bool {
	return l.requests.Enqueue(floor)
}

func (l *Lift) processFloorRequest() {
	if l.requests.Length() == 0 {
		return
	}
	floor := l.requests.Dequeue()

	l.moveToCalledFloor(floor)
	l.publish(LiftArrived{LiftId: l.Id, Floor: l.currentFloor})
	l.publish(LiftDoorsOpened{LiftId: l.Id, Floor: l.currentFloor})
	time.Sleep(time.Duration(l.doorCloseWaitMs) * time.Millisecond)
	l.publish(LiftDoorsClosed{LiftId: l.Id, Floor: l.currentFloor})
}

func (l *Lift) moveToCalledFloor(floor int) {
	var delta int
	if l.currentFloor > floor {
		delta = -1
	} else {
		delta = 1
	}
	for l.currentFloor != floor {
		l.transit(delta)
	}
}

func (l *Lift) transit(delta int) {
	curr := l.currentFloor
	// sleep for floorsPerSecond
	sleepTime := time.Second / time.Duration(l.floorsPerSecond)
	time.Sleep(sleepTime)
	l.currentFloor = l.currentFloor + delta
	l.publish(LiftTransited{LiftId: l.Id, From: curr, To: l.currentFloor})
}

// Start
// Gets the lift to listen for calls
func (l *Lift) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				l.processFloorRequest()
			}
		}
	}()
}

func (l *Lift) State() LiftState {
	return LiftState{
		Id:           l.Id,
		CurrentFloor: l.currentFloor,
		LowestFloor:  l.lowestFloor,
		HighestFloor: l.highestFloor,
	}
}

func (l *Lift) publish(event pubsub.Message) {
	l.pubsub.Publish(topic(l.Id), event)
}

func (l *Lift) Subscribe() (uuid.UUID, <-chan pubsub.Message, error) {
	return l.pubsub.Subscribe(topic(l.Id))
}

func (l *Lift) Unsubscribe(id uuid.UUID) {
	l.pubsub.Unsubscribe(id)
}

func (l *Lift) Call(floor int) bool {
	enqueued := l.enqueue(floor)
	if enqueued {
		l.pubsub.Publish(topic(l.Id), LiftCalled{LiftId: l.Id, Floor: floor})
	}
	return enqueued
}
