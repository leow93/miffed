package pubsub

type Message interface{}
type Topic string

type PubSub interface {
	Publish(topic Topic, message Message) error
	Subscribe(topic Topic) (<-chan Message, error)
}

type MemoryPubSub struct {
	subscribers map[Topic][]chan Message
}

func (ps *MemoryPubSub) Publish(topic Topic, message Message) error {
	for _, subscriber := range ps.subscribers[topic] {
		go func(subscriber chan Message) {
			subscriber <- message
		}(subscriber)
	}
	return nil
}

func (ps *MemoryPubSub) Subscribe(topic Topic) (<-chan Message, error) {
	subscriber := make(chan Message)
	ps.subscribers[topic] = append(ps.subscribers[topic], subscriber)
	return subscriber, nil
}

func NewMemoryPubSub() *MemoryPubSub {
	return &MemoryPubSub{
		subscribers: make(map[Topic][]chan Message),
	}
}
