package lift

import (
	"sync"

	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/pubsub"
)

const LiftRepoTopic = "lifts"

type LiftRepo struct {
	lifts  map[Id]*Lift
	pubsub pubsub.PubSub
	mutex  sync.Mutex
}

func NewLiftRepo(ps pubsub.PubSub) *LiftRepo {
	return &LiftRepo{
		lifts:  make(map[Id]*Lift),
		pubsub: ps,
		mutex:  sync.Mutex{},
	}
}

func (r *LiftRepo) AddLift(lift NewLiftOpts) *Lift {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	l := NewLift(r.pubsub, lift)
	r.lifts[l.Id] = l
	r.pubsub.Publish(LiftRepoTopic, LiftAdded{LiftId: l.Id, Lift: l})
	return l
}

func (r *LiftRepo) GetLift(id Id) *Lift {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.lifts[id]
}

func (r *LiftRepo) GetLifts() []Lift {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var lifts []Lift
	for _, l := range r.lifts {
		lifts = append(lifts, *l)
	}

	return lifts
}

func (r *LiftRepo) DeleteLift(id Id) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.lifts[id]; ok {
		delete(r.lifts, id)
		r.pubsub.Publish(LiftRepoTopic, LiftDeleted{LiftId: id})
	}
}

func (r *LiftRepo) Subscribe() (uuid.UUID, <-chan pubsub.Message, error) {
	return r.pubsub.Subscribe(LiftRepoTopic)
}

func (r *LiftRepo) Unsubscribe(id uuid.UUID) {
	r.pubsub.Unsubscribe(id)
}
