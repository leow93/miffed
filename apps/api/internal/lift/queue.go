package lift

import (
	"errors"
	"slices"
	"sync"
)

type Q interface {
	Enqueue(floor int) bool
	Dequeue() (int, error)
	Length() int
}

type Queue struct {
	queue []int
	mutex sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{}
}
func (q *Queue) Enqueue(floor int) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if slices.Contains(q.queue, floor) {
		return false
	}
	q.queue = append(q.queue, floor)
	return true
}
func (q *Queue) Dequeue() (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.queue) == 0 {
		// FIXME: assumes no negative floors
		return 0, errors.New("queue is empty")
	}
	floor := q.queue[0]
	q.queue = q.queue[1:]
	return floor, nil
}
func (q *Queue) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.queue)
}

type PriorityQueue struct {
	queue []int
	mutex sync.Mutex
}
