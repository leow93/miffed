package eventstore

type (
	Evolve[State any, Event any]      func(state State, event Event) State
	Decide[Command, State, Event any] func(cmd Command, state State) []Event
	InitialState[State any]           func() State
	StreamId[Command any]             func(cmd Command) string
	Deserialise[E any]                func(ev Event) (E, error)
	Serialise[E any]                  func(ev E) (Event, error)
)

type Decider[C, E, S any] struct {
	initialState S
	evolve       Evolve[S, E]
	decide       Decide[C, S, E]
	streamId     StreamId[C]
}

type eventStore interface {
	AppendToStream(streamName string, expectedVersion uint64, events []Event) error
	ReadStream(streamName string) ([]Event, uint64, error)
}

type DecisionFunc[C, E any] func(cmd C) error

func NewDecider[C, E, S any](
	store eventStore,
	initialState S,
	evolve Evolve[S, E],
	decide Decide[C, S, E],
	streamId StreamId[C],
	deserialise Deserialise[E],
	serialise Serialise[E],
) DecisionFunc[C, E] {
	return func(cmd C) error {
		streamName := streamId(cmd)
		rawEvents, version, err := store.ReadStream(streamName)
		if err != nil {
			return err
		}

		var domainEvents []E
		for _, e := range rawEvents {
			ev, err := deserialise(e)
			if err == nil {
				domainEvents = append(domainEvents, ev)
			}
		}

		state := initialState
		for _, ev := range domainEvents {
			state = evolve(state, ev)
		}
		newEvents := decide(cmd, state)

		if len(newEvents) > 0 {
			var serialisedEvs []Event
			for _, ev := range newEvents {
				serialised, err := serialise(ev)
				if err != nil {
					return err
				}
				serialisedEvs = append(serialisedEvs, serialised)
			}
			return store.AppendToStream(streamName, version, serialisedEvs)
		}

		return nil
	}
}
