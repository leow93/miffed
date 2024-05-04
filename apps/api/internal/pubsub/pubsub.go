package pubsub

import (
	"github.com/google/uuid"
	"sync"
)

type Message interface{}
type Topic string
type Unsubscribe func()

type PubSub interface {
	Publish(topic Topic, message Message) error
	Subscribe(topic Topic) (uuid.UUID, <-chan Message, error)
	Unsubscribe(id uuid.UUID)
}

type MemoryPubSub struct {
	subscribers map[Topic]map[uuid.UUID]chan Message
	mutex       sync.Mutex
}

func (ps *MemoryPubSub) addSubscriber(topic Topic, id uuid.UUID) <-chan Message {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	if _, ok := ps.subscribers[topic]; !ok {
		ps.subscribers[topic] = make(map[uuid.UUID]chan Message)
	}
	ch := make(chan Message)
	ps.subscribers[topic][id] = ch
	return ch
}

func (ps *MemoryPubSub) Publish(topic Topic, message Message) error {
	for _, subscriber := range ps.subscribers[topic] {
		go func(subscriber chan Message) {
			subscriber <- message
		}(subscriber)
	}
	return nil
}

func (ps *MemoryPubSub) Subscribe(topic Topic) (uuid.UUID, <-chan Message, error) {
	id := uuid.New()
	ch := ps.addSubscriber(topic, id)
	return id, ch, nil
}

func (ps *MemoryPubSub) Unsubscribe(id uuid.UUID) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	for _, subscriber := range ps.subscribers {
		if ch, ok := subscriber[id]; ok {
			close(ch)
			delete(subscriber, id)
		}
	}
}

func NewMemoryPubSub() *MemoryPubSub {
	return &MemoryPubSub{
		subscribers: make(map[Topic]map[uuid.UUID]chan Message),
	}
}
