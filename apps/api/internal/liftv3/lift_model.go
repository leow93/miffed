package liftv3

import (
	"context"
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
			fmt.Printf("Lift %s got a call to floor: %d\n", lift.Id.String(), floor)
			if lift.Floor == floor {
				continue
			}
			if !lift.floorsToVisit.Has(floor) {
				lift.floorsToVisit.Enqueue(floor)
				fmt.Printf("Lift %s will visit floor: %d\n", lift.Id.String(), floor)
			} else {
				fmt.Printf("Lift %s already going to visit to floor: %d\n", lift.Id.String(), floor)
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
				fmt.Printf("sending to transit chan: %d\n", nextFloor)
				lift.transitChan <- nextFloor
				fmt.Printf("sent to transit chan: %d\n", nextFloor)
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
				fmt.Printf("receive from transit chan: %d\n", nextFloor)
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
					fmt.Printf("%s moved from %d to %d on way to %d\n", lift.Id.String(), from, to, nextFloor)
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
				fmt.Printf("%s arrived at %d\n", lift.Id.String(), lift.Floor)
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

type (
	SubscriptionId int
	ILiftService   interface {
		AddLift(ctx context.Context, cfg LiftConfig) (Lift, error)
		GetLift(ctx context.Context, id LiftId) (Lift, error)
		CallLift(ctx context.Context, id LiftId, floor int) error
	}
)

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

func liftNotFoundErr(id LiftId) error {
	return fmt.Errorf("lift not found: %s", id.String())
}

func (svc *LiftService) getLiftModel(id LiftId) (*liftModel, error) {
	svc.mx.Lock()
	defer svc.mx.Unlock()
	lift, ok := svc.lifts[id]
	if !ok {
		return &liftModel{}, liftNotFoundErr(id)
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
	go lift.handleNotifications(ctx, svc.Notifications)
}
