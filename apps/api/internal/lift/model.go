package lift

import (
	"context"
	"log"
	"time"
)

type Lift struct {
	lowestFloor   int
	highestFloor  int
	currentFloor  int
	speed         int    // queue per second
	requests      *Queue // queue to visit
	ctx           context.Context
	subscriptions []chan Event
}

func NewLift(ctx context.Context, lowestFloor, highestFloor, currentFloor, floorsPerSecond int) *Lift {
	lift := &Lift{
		lowestFloor:   lowestFloor,
		highestFloor:  highestFloor,
		currentFloor:  currentFloor,
		speed:         floorsPerSecond,
		requests:      NewQueue(),
		ctx:           ctx,
		subscriptions: make([]chan Event, 0),
	}
	lift.Start()
	return lift
}

func (l *Lift) enqueue(floor int) bool {
	return l.requests.Enqueue(floor)
}

// TODO: processor should probably be a separate thing
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
	l.arrive()
}

// FIXME: this is not the responsibility of the lift
func (l *Lift) notify(event Event) {
	logger := log.Default()
	switch event.(type) {
	case LiftArrived:
		logger.Println("Lift arrived at floor", event.(LiftArrived).Floor)
	case LiftCalled:
		logger.Println("Lift called to floor", event.(LiftCalled).Floor)
	case LiftTransited:
		logger.Println("Lift transited from floor", event.(LiftTransited).From, "to floor", event.(LiftTransited).To)
	}

	for _, sub := range l.subscriptions {
		sub <- event
	}
}

func (l *Lift) transit(delta int) {
	curr := l.currentFloor
	// sleep for speed
	sleepTime := time.Second / time.Duration(l.speed)
	time.Sleep(sleepTime)
	l.currentFloor = l.currentFloor + delta
	l.notify(LiftTransited{From: curr, To: l.currentFloor})
}

func (l *Lift) arrive() {
	l.notify(LiftArrived{Floor: l.currentFloor})
}

// Start
// Gets the lift to listen for calls
func (l *Lift) Start() {
	go func() {
		for {
			select {
			case <-l.ctx.Done():
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

func (l *Lift) Subscribe() chan Event {
	sub := make(chan Event)
	l.subscriptions = append(l.subscriptions, sub)
	return sub
}

func (l *Lift) Call(floor int) {
	go func() {
		enqueued := l.enqueue(floor)
		if enqueued {
			l.notify(LiftCalled{Floor: floor})
		}
	}()
}
