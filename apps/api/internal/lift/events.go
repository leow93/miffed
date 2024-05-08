package lift

type LiftCalled struct {
	LiftId Id  `json:"liftId"`
	Floor  int `json:"floor"`
}
type LiftTransited struct {
	LiftId Id  `json:"liftId"`
	From   int `json:"from"`
	To     int `json:"to"`
}
type LiftArrived struct {
	LiftId Id  `json:"liftId"`
	Floor  int `json:"floor"`
}
type LiftDoorsOpened struct {
	LiftId Id  `json:"liftId"`
	Floor  int `json:"floor"`
}
type Event interface{}

type LiftMessage struct {
	LiftId Id          `json:"liftId"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
}

func SerialiseEvent(e Event) *LiftMessage {
	switch e.(type) {
	case LiftArrived:
		return &LiftMessage{
			Type:   "lift_arrived",
			Data:   e,
			LiftId: e.(LiftArrived).LiftId,
		}
	case LiftCalled:
		return &LiftMessage{
			Type:   "lift_called",
			Data:   e,
			LiftId: e.(LiftCalled).LiftId,
		}
	case LiftTransited:
		return &LiftMessage{
			Type:   "lift_transited",
			Data:   e,
			LiftId: e.(LiftTransited).LiftId,
		}

	case LiftDoorsOpened:
		return &LiftMessage{
			Type:   "lift_doors_opened",
			Data:   e,
			LiftId: e.(LiftDoorsOpened).LiftId,
		}
	default:
		return nil
	}
}
