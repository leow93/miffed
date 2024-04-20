package pubsub

import "testing"

func TestMemoryPubSub(t *testing.T) {
	t.Run("publishing a message", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		message := "hello"
		err := pubsub.Publish("test", message)
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}
	})

	t.Run("subscribing to a topic", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		subscription, err := pubsub.Subscribe("test")
		pubsub.Publish("test", "hello")
		if err != nil {
			t.Errorf("got error %v, want nil", err)
		}
		message := <-subscription
		if message != "hello" {
			t.Errorf("got message %v, want hello", message)
		}
	})
	t.Run("multiple subscribers receive the message", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		subscription1, _ := pubsub.Subscribe("test")
		subscription2, _ := pubsub.Subscribe("test")
		pubsub.Publish("test", "hello")
		message1 := <-subscription1
		message2 := <-subscription2
		if message1 != "hello" {
			t.Errorf("got message %v, want hello", message1)
		}
		if message2 != "hello" {
			t.Errorf("got message %v, want hello", message2)
		}
	})
	t.Run("subscriptions only receive messages for their topic", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		subscription1, _ := pubsub.Subscribe("foo")
		subscription2, _ := pubsub.Subscribe("bar")
		pubsub.Publish("foo", "hello")
		pubsub.Publish("bar", "world")
		message1 := <-subscription1
		message2 := <-subscription2
		if message1 != "hello" {
			t.Errorf("got message %v, want hello", message1)
		}
		if message2 != "world" {
			t.Errorf("got message %v, want world", message2)
		}
	})
	t.Run("subscriptions only receive messages once they have subscribed", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		pubsub.Publish("foo", "hello")
		subscription1, _ := pubsub.Subscribe("foo")
		pubsub.Publish("foo", "world")
		message1 := <-subscription1
		if message1 != "world" {
			t.Errorf("got message %v, want world", message1)
		}
	})
}
