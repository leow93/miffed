package lift

import (
	"context"
	"errors"
	"sync"

	"github.com/leow93/miffed-api/internal/pubsub"
)

type LiftSub struct {
	unsubscribe func()
}

type Aggregator struct {
	mutex          sync.Mutex
	liftSubs       map[Id]LiftSub
	ps             pubsub.PubSub
	repoChan       chan LiftMessage
	liftMotionChan chan LiftMessage
	errChan        chan error
	sinkChan       chan LiftMessage
}

func NewAggregator(ctx context.Context, ps pubsub.PubSub, repo *LiftRepo, sinkChan chan LiftMessage, errChan chan error) *Aggregator {
	lifts := repo.GetLifts()

	aggregator := Aggregator{
		mutex:          sync.Mutex{},
		liftSubs:       make(map[Id]LiftSub),
		ps:             ps,
		repoChan:       make(chan LiftMessage),
		liftMotionChan: make(chan LiftMessage),
		errChan:        errChan,
		sinkChan:       sinkChan,
	}

	for _, l := range lifts {
		aggregator.subscribeToLiftMotion(ctx, &l)
	}

	// Readiness WG allows our subscriptions to start up before we return the aggregator to the caller
	readiness := sync.WaitGroup{}
	readiness.Add(2)
	go aggregator.subscribeToRepo(ctx, repo, &readiness)
	go aggregator.startSink(ctx, &readiness)
	readiness.Wait()

	return &aggregator
}

func (agg *Aggregator) subscribeToLiftMotion(ctx context.Context, lift *Lift) {
	if lift == nil {
		agg.errChan <- errors.New("received nil pointer for lift")
		return
	}
	agg.mutex.Lock()
	defer agg.mutex.Unlock()
	id, ch, err := lift.Subscribe()
	if err != nil {
		agg.errChan <- err
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:

				ev := SerialiseEvent(msg)

				if ev == nil {
					continue
				}
				agg.liftMotionChan <- *ev
			}
		}
	}()

	sub := LiftSub{
		unsubscribe: func() {
			lift.Unsubscribe(id)
		},
	}
	agg.liftSubs[lift.Id] = sub
}

func (agg *Aggregator) unsubscribeFromLiftMotion(id Id) {
	agg.mutex.Lock()
	defer agg.mutex.Unlock()

	if sub, ok := agg.liftSubs[id]; ok {
		sub.unsubscribe()
	}
}

func (agg *Aggregator) subscribeToRepo(ctx context.Context, repo *LiftRepo, readiness *sync.WaitGroup) {
	id, ch, err := repo.Subscribe()
	if err != nil {
		agg.errChan <- err
		return
	}
	readiness.Done()

	for {
		select {
		case <-ctx.Done():
			repo.Unsubscribe(id)
			return
		case msg := <-ch:
			e := SerialiseEvent(msg)
			if e == nil {
				continue
			}
			if e.Type == "lift_added" {
				data := e.Data.(LiftAdded)
				agg.subscribeToLiftMotion(ctx, data.Lift)
				agg.sinkChan <- *e
			} else if e.Type == "lift_deleted" {
				agg.unsubscribeFromLiftMotion(e.LiftId)
				agg.sinkChan <- *e
			}
		}
	}
}

func (agg *Aggregator) startSink(ctx context.Context, readiness *sync.WaitGroup) {
	readiness.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-agg.liftMotionChan:
			agg.sinkChan <- msg
		}
	}
}
