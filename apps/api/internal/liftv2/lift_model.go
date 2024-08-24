package liftv2

// 1. The service is a layer that interacts with individual lifts, allowing
// clients to create and call lifts.

// 2. The notifier allows clients to subscribe to lift events: creation, movement etc.

type LiftConfig struct {
	Floor int
}

// TODO: rename
type Lift2 struct {
	Id    LiftId
	Floor int
}
type (
	SubscriptionId int
	ILiftService   interface {
		AddLift(cfg LiftConfig) error
		GetLift(id LiftId) (Lift2, error)
		CallLift(id LiftId, floor int) error
		Subscribe(receiver chan<- LiftEvent) (SubscriptionId, error)
		Unsubscribe(id SubscriptionId) error
	}
)
