package liftv2

import (
	"context"
	"sync"
	"time"

	"github.com/leow93/miffed-api/internal/eventstore"
)

type categoryReader interface {
	ReadCategory(categoryName string, fromPosition uint64) ([]eventstore.Event, error)
}

type LiftReadModel struct {
	mutex sync.Mutex
	lifts map[LiftId]LiftModel
	store categoryReader
}

func NewLiftReadModel(ctx context.Context, store categoryReader) *LiftReadModel {
	model := &LiftReadModel{
		mutex: sync.Mutex{},
		store: store,
		lifts: make(map[LiftId]LiftModel),
	}

	go model.startSubscription(ctx)

	return model
}

func (lrm *LiftReadModel) Query() []LiftModel {
	lrm.mutex.Lock()
	defer lrm.mutex.Unlock()
	var result []LiftModel
	for _, lift := range lrm.lifts {
		result = append(result, lift)
	}
	return result
}

func (lrm *LiftReadModel) startSubscription(ctx context.Context) {
	checkpoint := uint64(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			batch, err := lrm.getNextBatch(&checkpoint)
			if err != nil {
				continue
			}
			lrm.handleBatch(batch)

		}
	}
}

func (lrm *LiftReadModel) getNextBatch(checkpoint *uint64) ([]LiftEvent, error) {
	rawEvs, err := lrm.store.ReadCategory("Lift", *checkpoint)
	if err != nil {
		return []LiftEvent{}, err
	}

	var evs []LiftEvent
	for _, rawEv := range rawEvs {
		ev, err := deserialise(rawEv)
		if err == nil {
			evs = append(evs, ev)
		}
	}

	newCheckpoint := uint64(len(evs))
	*checkpoint = newCheckpoint

	return evs, err
}

func (lrm *LiftReadModel) handleBatch(evs []LiftEvent) {
	for _, ev := range evs {
		if ev.eventType() == "lift_added" {
			lrm.mutex.Lock()
			data := ev.(LiftAdded)

			lrm.lifts[data.Id] = LiftModel{
				Id:    data.Id,
				Floor: data.Floor,
				Speed: data.Speed,
			}

			lrm.mutex.Unlock()
		}
	}
}
