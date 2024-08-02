package lift

import (
	"sync"

	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/pubsub"
)

type Manager struct {
	pubsub                        pubsub.PubSub
	lifts                         map[Id]*Lift
	globalToLiftSubscriptionIdMap map[uuid.UUID][]uuid.UUID
	mutex                         sync.Mutex
}

type ManagerState map[Id]LiftState

func NewManager(ps pubsub.PubSub) *Manager {
	return &Manager{
		lifts:                         make(map[Id]*Lift),
		pubsub:                        ps,
		globalToLiftSubscriptionIdMap: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *Manager) AddLift(opts NewLiftOpts) *Lift {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	lift := NewLift(m.pubsub, opts)
	m.lifts[lift.Id] = lift
	return lift
}

// todo: func(m *Manager) StartLift(id Id) {}
func (m *Manager) GetLift(id Id) *Lift {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.lifts[id]
}

func (m *Manager) CallLift(id Id, floor int) bool {
	l := m.GetLift(id)
	if l == nil {
		return false
	}
	return l.Call(floor)
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
		liftSubId, liftChan, e := l.Subscribe()
		if e != nil {
			m.pubsub.Unsubscribe(liftSubId)
			return id, ch, e
		}

		m.globalToLiftSubscriptionIdMap[id] = append(m.globalToLiftSubscriptionIdMap[id], liftSubId)

		go func(channel <-chan pubsub.Message) {
			for {
				msg := <-channel
				m.pubsub.Publish("lift", msg)
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

func (m *Manager) State() ManagerState {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	state := make(ManagerState)
	for id, l := range m.lifts {
		state[id] = l.State()
	}
	return state
}
