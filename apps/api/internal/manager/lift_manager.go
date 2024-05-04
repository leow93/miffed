package manager

import (
	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/lift"
	"github.com/leow93/miffed-api/internal/pubsub"
	"sync"
)

type Manager struct {
	pubsub                        pubsub.PubSub
	lifts                         map[lift.Id]*lift.Lift
	globalToLiftSubscriptionIdMap map[uuid.UUID][]uuid.UUID
	mutex                         sync.Mutex
}

func NewManager(ps pubsub.PubSub) *Manager {
	return &Manager{
		lifts:                         make(map[lift.Id]*lift.Lift),
		pubsub:                        ps,
		globalToLiftSubscriptionIdMap: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *Manager) AddLift(l *lift.Lift) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.lifts[l.Id] = l
}

func (m *Manager) GetLift(id lift.Id) *lift.Lift {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.lifts[id]
}

func (m *Manager) CallLift(id lift.Id, floor int) {
	l := m.GetLift(id)
	if l == nil {
		return
	}
	l.Call(floor)
}

func (m *Manager) Subscribe() (uuid.UUID, <-chan pubsub.Message, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	id, ch, err := m.pubsub.Subscribe("lift")
	if err != nil {
		return id, ch, err
	}

	// pump messages from individual lift topics to the global lift topic
	for _, l := range m.lifts {
		liftSubId, liftChan, e := m.pubsub.Subscribe(lift.Topic(l.Id))
		if e != nil {
			m.pubsub.Unsubscribe(liftSubId)
			return id, ch, e
		}

		m.globalToLiftSubscriptionIdMap[id] = append(m.globalToLiftSubscriptionIdMap[id], liftSubId)

		go func(channel <-chan pubsub.Message) {
			for {
				m.pubsub.Publish("lift", <-channel)
			}
		}(liftChan)
	}

	return id, ch, nil
}

func (m *Manager) Unsubscribe(id uuid.UUID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.pubsub.Unsubscribe(id)
	ids, ok := m.globalToLiftSubscriptionIdMap[id]
	if !ok {
		return
	}
	for _, internalId := range ids {
		m.pubsub.Unsubscribe(internalId)
	}
	delete(m.globalToLiftSubscriptionIdMap, id)
}
