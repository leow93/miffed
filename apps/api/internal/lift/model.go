package lift

import (
	"context"
	"slices"
	"time"
)

type Queue struct {
	queue []int
}

func NewQueue() *Queue {
	return &Queue{}
}
func (q *Queue) Enqueue(floor int) {
	if slices.Contains(q.queue, floor) {
		return
	}
	q.queue = append(q.queue, floor)
}
func (q *Queue) Dequeue() int {
	if len(q.queue) == 0 {
		// FIXME: assumes no negative floors
		return -1
	}
	floor := q.queue[0]
	q.queue = q.queue[1:]
	return floor
}
func (q *Queue) Length() int {
	return len(q.queue)
}

type Lift struct {
	lowestFloor  int
	highestFloor int
	currentFloor int
	speed        int      // queue per second
	requests     *Queue   // queue to visit
	arrival      chan int // floor the lift has arrived at
	ctx          context.Context
}

func NewLift(lowestFloor, highestFloor, currentFloor, floorsPerSecond int) *Lift {
	lift := &Lift{
		lowestFloor:  lowestFloor,
		highestFloor: highestFloor,
		currentFloor: currentFloor,
		speed:        floorsPerSecond,
		requests:     NewQueue(),
		arrival:      make(chan int),
	}
	lift.Start()
	return lift
}

func (l *Lift) enqueue(floor int) {
	l.requests.Enqueue(floor)
}

func (l *Lift) waitForArrival(floor int, done chan bool) {
	// wait for the lift to arrive
	f := <-l.arrival
	if f != floor {
		done <- false
	} else {
		done <- true
	}
}

// Start
// Gets the lift to listen for calls
func (l *Lift) Start() {
	// take off the queue and move the lift there according to speed
	// send the result to the arrival channel
	go func() {
		for {
			if l.requests.Length() == 0 {
				continue
			}
			floor := l.requests.Dequeue()
			var distance int
			if l.currentFloor > floor {
				distance = l.currentFloor - floor
			} else {
				distance = floor - l.currentFloor
			}
			// move the lift there according to speed
			for i := 0; i < distance; i++ {
				l.currentFloor += 1
				// sleep for speed
				sleepTime := time.Second / time.Duration(l.speed)
				time.Sleep(sleepTime)
			}
			// send the result to the arrival channel
			l.arrival <- l.currentFloor
		}
	}()
}

func (l *Lift) Call(floor int) <-chan bool {
	done := make(chan bool, 1)
	go func() {
		l.enqueue(floor)
		l.waitForArrival(floor, done)
	}()
	return done
}
