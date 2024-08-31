package lift

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/leow93/miffed-api/internal/pubsub"
	"github.com/leow93/miffed-api/internal/queue"
)

type LiftConfig struct {
	Floor        int
	FloorDelayMs int
}

type Lift struct {
	Id           LiftId
	Floor        int
	floorDelayMs int
}

type liftModel struct {
	Lift
	floorsToVisit *queue.Queue
	callsChan     chan int       // channel which buffers client calls
	transitChan   chan int       // channel which takes valid floors to visit and moves there one by one
	notifications chan LiftEvent // channel for clients to receive notifications on
	floorDelayMs  int
	mx            sync.RWMutex
}

func newLiftModel(lift Lift) *liftModel {
	return &liftModel{
		Lift:          lift,
		floorsToVisit: queue.NewQueue(),
		callsChan:     make(chan int),
		transitChan:   make(chan int),
		notifications: make(chan LiftEvent),
		floorDelayMs:  lift.floorDelayMs,
		mx:            sync.RWMutex{},
	}
}

func (lift *liftModel) currentFloor() int {
	lift.mx.RLock()
	defer lift.mx.RUnlock()
	return lift.Floor
}

func (lift *liftModel) transitToFloor(ctx context.Context, delta int) {
	lift.mx.Lock()
	defer lift.mx.Unlock()
	from := lift.Floor
	to := lift.Floor + delta
	lift.Floor = to
	lift.publish(ctx, createLiftEvent(lift.Id, "lift_transited", LiftTransited{From: from, To: to}))
	time.Sleep(time.Duration(lift.floorDelayMs * 1000 * 1000))
}

func (lift *liftModel) call(ctx context.Context, floor int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return fmt.Errorf("timed out calling lift")
	case lift.callsChan <- floor:
		return nil
	}
}

func (lift *liftModel) publish(ctx context.Context, ev LiftEvent) {
	select {
	case <-ctx.Done():
		return
	case lift.notifications <- ev:
		return
	}
}

func (lift *liftModel) handleCalls(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case floor := <-lift.callsChan:
			if lift.currentFloor() == floor {
				continue
			}
			if !lift.floorsToVisit.Has(floor) {
				lift.floorsToVisit.Enqueue(floor)
			}
		}
	}
}

func (lift *liftModel) handleFloorsToVisit(ctx context.Context) {
	// Send floors to visit to transit chan
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				nextFloor, err := lift.floorsToVisit.Dequeue()
				if err != nil {
					continue
				}
				lift.transitChan <- nextFloor
			}
		}
	}()

	// Send floors to visit to transit chan
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case nextFloor := <-lift.transitChan:
				var delta int
				if lift.currentFloor() > nextFloor {
					delta = -1
				} else {
					delta = 1
				}
				for lift.currentFloor() != nextFloor {
					lift.transitToFloor(ctx, delta)
				}

				lift.publish(ctx, createLiftEvent(lift.Id, "lift_arrived", LiftArrived{Floor: nextFloor}))
			}
		}
	}()
}

func (lift *liftModel) handleNotifications(ctx context.Context, publish publish) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-lift.notifications:
			publish(ev)
		}
	}
}

type subscriptionId struct {
	uuid.UUID
}

type subscription struct {
	Id       subscriptionId
	EventsCh chan LiftEvent
	ctx      context.Context
	cancel   context.CancelFunc
}

func newSubscription(ctx context.Context, cancel context.CancelFunc) *subscription {
	return &subscription{
		Id:       subscriptionId{uuid.New()},
		EventsCh: make(chan LiftEvent),
		ctx:      ctx,
		cancel:   cancel,
	}
}

type publish func(ev any) error

type LiftService struct {
	liftOrder     []LiftId
	lifts         map[LiftId]*liftModel
	mx            sync.Mutex
	lifecycleChan chan *liftModel
	notifications chan LiftEvent
	publish       publish
}

func NewLiftService(ctx context.Context, ps pubsub.PubSub) *LiftService {
	publish := func(ev any) error {
		return ps.Publish("lifts", ev)
	}
	svc := &LiftService{
		lifts:         make(map[LiftId]*liftModel),
		mx:            sync.Mutex{},
		lifecycleChan: make(chan *liftModel),
		notifications: make(chan LiftEvent),
		publish:       publish,
	}
	go svc.manageLiftLifecycle(ctx)
	return svc
}

func (svc *LiftService) AddLift(cfg LiftConfig) (Lift, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	id := NewLiftId()
	lift := Lift{
		Id:           id,
		Floor:        cfg.Floor,
		floorDelayMs: cfg.FloorDelayMs,
	}
	liftModel := newLiftModel(lift)
	svc.lifts[id] = liftModel
	svc.liftOrder = append(svc.liftOrder, id)
	go func() {
		svc.lifecycleChan <- liftModel
	}()
	go func() {
		liftModel.notifications <- createLiftEvent(liftModel.Id, "lift_added", LiftAdded{Floor: liftModel.currentFloor()})
	}()
	return lift, nil
}

var ErrLiftNotFound = errors.New("lift not found")

func (svc *LiftService) getLiftModel(id LiftId) (*liftModel, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	lift, ok := svc.lifts[id]
	if !ok {
		return &liftModel{}, ErrLiftNotFound
	}
	return lift, nil
}

func (svc *LiftService) GetLift(_ context.Context, id LiftId) (Lift, error) {
	model, err := svc.getLiftModel(id)
	if err != nil {
		return Lift{}, err
	}

	return model.Lift, nil
}

func (svc *LiftService) GetLifts(_ context.Context) ([]Lift, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	result := make([]Lift, len(svc.liftOrder))

	for i, id := range svc.liftOrder {
		lift, ok := svc.lifts[id]
		if !ok {
			continue
		}
		result[i] = Lift{Id: lift.Id, Floor: lift.currentFloor()}
	}

	return result, nil
}

func (svc *LiftService) CallLift(ctx context.Context, id LiftId, floor int) error {
	model, err := svc.getLiftModel(id)
	if err != nil {
		return err
	}

	return model.call(ctx, floor)
}

func (svc *LiftService) manageLiftLifecycle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case lift := <-svc.lifecycleChan:
			svc.startLiftProcessing(ctx, lift)
		}
	}
}

func (svc *LiftService) startLiftProcessing(ctx context.Context, lift *liftModel) {
	go lift.handleCalls(ctx)
	go lift.handleFloorsToVisit(ctx)
	go lift.handleNotifications(ctx, svc.publish)
}

type SubscriptionManager struct {
	pubsub pubsub.PubSub
}

func NewSubscriptionManager(backgroundCtx context.Context, ps pubsub.PubSub) *SubscriptionManager {
	return &SubscriptionManager{
		pubsub: ps,
	}
}

func (s *SubscriptionManager) Subscribe() (uuid.UUID, <-chan LiftEvent, error) {
	id, ch, err := s.pubsub.Subscribe("lifts")
	eventsCh := make(chan LiftEvent)
	if err != nil {
		return id, eventsCh, err
	}
	go func() {
		for {
			// TODO: make pubsub generic
			msg := <-ch
			event, ok := msg.(LiftEvent)
			if !ok {
				continue
			}
			eventsCh <- event
		}
	}()

	return id, eventsCh, nil
}

func (s *SubscriptionManager) Unsubscribe(id uuid.UUID) error {
	s.pubsub.Unsubscribe(id)
	return nil
}
