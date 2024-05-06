package pubsub

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Message interface{}
type Topic string

type PubSub interface {
	Publish(topic Topic, message Message) error
	Subscribe(topic Topic) (uuid.UUID, <-chan Message, error)
	Unsubscribe(id uuid.UUID)
}

type subscriber struct {
	ch     chan Message
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

type MemoryPubSub struct {
	subscribers map[Topic]map[uuid.UUID]subscriber
	mutex       sync.Mutex
}

func (ps *MemoryPubSub) addSubscriber(topic Topic, id uuid.UUID) <-chan Message {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	ctx, cncl := context.WithCancel(context.Background())
	if _, ok := ps.subscribers[topic]; !ok {
		ps.subscribers[topic] = make(map[uuid.UUID]subscriber)
	}
	ch := make(chan Message)
	ps.subscribers[topic][id] = subscriber{
		ch:     ch,
		ctx:    ctx,
		cancel: cncl,
		once:   sync.Once{},
	}
	return ch
}

func (ps *MemoryPubSub) Publish(topic Topic, message Message) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	for _, s := range ps.subscribers[topic] {
		go func(s subscriber) {
			select {
			case <-s.ctx.Done():
				return
			case s.ch <- message:
			}
		}(s)
	}
	return nil
}

func (ps *MemoryPubSub) Subscribe(topic Topic) (uuid.UUID, <-chan Message, error) {
	id := uuid.New()
	ch := ps.addSubscriber(topic, id)

	var subscribers []string
	for _, s := range ps.subscribers {
		for k := range s {
			subscribers = append(subscribers, k.String())
		}
	}

	fmt.Println("Subscribers", subscribers)

	return id, ch, nil
}

func (ps *MemoryPubSub) Unsubscribe(id uuid.UUID) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()
	for _, subs := range ps.subscribers {
		if sub, ok := subs[id]; ok {
			sub.cancel()
			delete(subs, id)
		}
	}
}

func NewMemoryPubSub() *MemoryPubSub {
	return &MemoryPubSub{
		subscribers: make(map[Topic]map[uuid.UUID]subscriber),
	}
}
