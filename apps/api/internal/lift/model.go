package lift

import (
	"context"
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
	Id           Id
	lowestFloor  int
	highestFloor int
	currentFloor int
	speed        int    // queue per second
	requests     *Queue // queue to visit
	pubsub       pubsub.PubSub
}

type LiftState struct {
	Id           Id  `json:"id"`
	CurrentFloor int `json:"currentFloor"`
	LowestFloor  int `json:"lowestFloor"`
	HighestFloor int `json:"highestFloor"`
}

func Topic(liftId Id) pubsub.Topic {
	return pubsub.Topic("lift:" + strconv.Itoa(int(liftId)))
}

func NewLift(ps pubsub.PubSub, lowestFloor, highestFloor, floorsPerSecond int) *Lift {
	lift := &Lift{
		Id:           NewId(),
		lowestFloor:  lowestFloor,
		highestFloor: highestFloor,
		currentFloor: lowestFloor,
		speed:        floorsPerSecond,
		requests:     NewQueue(),
		pubsub:       ps,
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

	l.transitionToFloor(floor)
}

func (l *Lift) transitionToFloor(floor int) {
	var delta int
	if l.currentFloor > floor {
		delta = -1
	} else {
		delta = 1
	}
	for l.currentFloor != floor {
		l.transit(delta)
	}
	l.pubsub.Publish(Topic(l.Id), LiftArrived{LiftId: l.Id, Floor: l.currentFloor})
}

func (l *Lift) transit(delta int) {
	curr := l.currentFloor
	// sleep for speed
	sleepTime := time.Second / time.Duration(l.speed)
	time.Sleep(sleepTime)
	l.currentFloor = l.currentFloor + delta
	l.pubsub.Publish(Topic(l.Id), LiftTransited{LiftId: l.Id, From: curr, To: l.currentFloor})
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

func (l *Lift) Call(floor int) bool {
	enqueued := l.enqueue(floor)
	if enqueued {
		l.pubsub.Publish(Topic(l.Id), LiftCalled{LiftId: l.Id, Floor: floor})
	}
	return enqueued
}
