package lift

import (
	"slices"
	"sync"
)

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
func (q *Queue) Dequeue() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.queue) == 0 {
		// FIXME: assumes no negative floors
		return -1
	}
	floor := q.queue[0]
	q.queue = q.queue[1:]
	return floor
}
func (q *Queue) Length() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.queue)
}
