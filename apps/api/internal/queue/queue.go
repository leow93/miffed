package queue

import (
	"errors"
	"sync"
)

type Queue struct {
	queue []int
	mutex sync.Mutex
}

func emptyQueue() error {
	return errors.New("queue is empty")
}

func NewQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Enqueue(value int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.queue = append(q.queue, value)
}

func (q *Queue) Dequeue() (int, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if len(q.queue) == 0 {
		return 0, emptyQueue()
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

func (q *Queue) Has(x int) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, y := range q.queue {
		if y == x {
			return true
		}
	}
	return false
}
