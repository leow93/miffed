package lift

type LiftCalled struct {
	Floor int `json:"floor"`
}
type LiftTransited struct {
	From int `json:"from"`
	To   int `json:"to"`
}
type LiftArrived struct {
	Floor int `json:"floor"`
}
type Event interface{}

type LiftMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func SerialiseEvent(e Event) *LiftMessage {
	switch e.(type) {
	case LiftArrived:
		return &LiftMessage{
			Type: "lift_arrived",
			Data: e,
		}
	case LiftCalled:
		return &LiftMessage{
			Type: "lift_called",
			Data: e,
		}
	case LiftTransited:
		return &LiftMessage{
			Type: "lift_transited",
			Data: e,
		}
	default:
		return nil
	}
}
