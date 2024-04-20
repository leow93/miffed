package lift

type LiftCalled struct {
	Floor int
}
type LiftTransited struct {
	From int
	To   int
}
type LiftArrived struct {
	Floor int
}
type Event interface{}
