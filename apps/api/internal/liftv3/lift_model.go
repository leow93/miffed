package liftv3

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/leow93/miffed-api/internal/queue"
)

type LiftConfig struct {
	Floor int
}

type Lift struct {
	Id    LiftId
	Floor int
}

type liftModel struct {
	Lift
	floorsToVisit *queue.Queue
	callsChan     chan int       // channel which buffers client calls
	transitChan   chan int       // channel which takes valid floors to visit and moves there one by one
	notifications chan LiftEvent // channel for clients to receive notifications on
}

func newLiftModel(lift Lift) *liftModel {
	return &liftModel{
		Lift:          lift,
		floorsToVisit: queue.NewQueue(),
		callsChan:     make(chan int),
		transitChan:   make(chan int),
		notifications: make(chan LiftEvent),
	}
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

func (lift *liftModel) handleCalls(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case floor := <-lift.callsChan:
			if lift.Floor == floor {
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
				if lift.Floor > nextFloor {
					delta = -1
				} else {
					delta = 1
				}
				evs := []LiftEvent{}
				for lift.Floor != nextFloor {
					from := lift.Floor
					to := lift.Floor + delta
					lift.Floor = to
					evs = append(evs, createLiftEvent(lift.Id, "lift_transited", LiftTransited{From: from, To: to}))
				}

				evs = append(evs, createLiftEvent(lift.Id, "lift_arrived", LiftArrived{Floor: nextFloor}))
				go func() {
					for _, ev := range evs {
						select {
						case <-ctx.Done():
							return
						default:
							lift.notifications <- ev
						}
					}
				}()
			}
		}
	}()
}

func (lift *liftModel) handleNotifications(ctx context.Context, dest chan<- LiftEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-lift.notifications:
			dest <- ev
		}
	}
}

type LiftService struct {
	lifts         map[LiftId]*liftModel
	mx            sync.Mutex
	lifecycleChan chan *liftModel
	Notifications chan LiftEvent
}

func NewLiftService(ctx context.Context) *LiftService {
	svc := &LiftService{
		lifts:         make(map[LiftId]*liftModel),
		mx:            sync.Mutex{},
		lifecycleChan: make(chan *liftModel),
		Notifications: make(chan LiftEvent),
	}
	go svc.manageLiftLifecycle(ctx)
	return svc
}

func (svc *LiftService) AddLift(cfg LiftConfig) (Lift, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	id := NewLiftId()
	lift := Lift{
		Id:    id,
		Floor: cfg.Floor,
	}
	liftModel := newLiftModel(lift)
	svc.lifts[id] = liftModel
	go func() {
		svc.lifecycleChan <- liftModel
	}()
	go func() {
		liftModel.notifications <- createLiftEvent(liftModel.Id, "lift_added", LiftAdded{Floor: liftModel.Floor})
	}()
	return lift, nil
}

var liftNotFoundErr = errors.New("lift not found")

func (svc *LiftService) getLiftModel(id LiftId) (*liftModel, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	lift, ok := svc.lifts[id]
	if !ok {
		return &liftModel{}, liftNotFoundErr
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
	result := make([]Lift, len(svc.lifts))

	i := 0
	for _, lift := range svc.lifts {
		result[i] = Lift{Id: lift.Id, Floor: lift.Floor}
		i++
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

// TODO: this should be the responsibility of a separate object
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
	go lift.handleNotifications(ctx, svc.Notifications)
}
