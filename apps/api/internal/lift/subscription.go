package lift

import "log"

func NewLiftSubscription(lift *Lift) (<-chan Event, error) {
	_, ch, err := lift.pubsub.Subscribe("lift")
	if err != nil {
		log.Println("Error subscribing to lift", err)
		return nil, err
	}
	evCh := make(chan Event)
	go func() {
		for {
			ev := <-ch
			evCh <- ev
		}
	}()
	return evCh, nil
}
