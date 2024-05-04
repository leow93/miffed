package lift

import (
	"context"
	"github.com/leow93/miffed-api/internal/pubsub"
	"time"
)

type Lift struct {
	lowestFloor  int
	highestFloor int
	currentFloor int
	speed        int    // queue per second
	requests     *Queue // queue to visit
	pubsub       pubsub.PubSub
}

func NewLift(ps pubsub.PubSub, lowestFloor, highestFloor, floorsPerSecond int) *Lift {
	lift := &Lift{
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
	l.pubsub.Publish("lift", LiftArrived{Floor: l.currentFloor})
}

func (l *Lift) transit(delta int) {
	curr := l.currentFloor
	// sleep for speed
	sleepTime := time.Second / time.Duration(l.speed)
	time.Sleep(sleepTime)
	l.currentFloor = l.currentFloor + delta
	l.pubsub.Publish("lift", LiftTransited{From: curr, To: l.currentFloor})
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

func (l *Lift) CurrentFloor() int {
	return l.currentFloor
}

func (l *Lift) HighestFloor() int {
	return l.highestFloor
}

func (l *Lift) LowestFloor() int {
	return l.lowestFloor
}

func (l *Lift) Call(floor int) bool {
	enqueued := l.enqueue(floor)
	if enqueued {
		l.pubsub.Publish("lift", LiftCalled{Floor: floor})
	}
	return enqueued
}
