package pubsub

import (
	"testing"
	"time"
)

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
		_, subscription, err := pubsub.Subscribe("test")
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
		_, subscription1, _ := pubsub.Subscribe("test")
		_, subscription2, _ := pubsub.Subscribe("test")
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
		_, subscription1, _ := pubsub.Subscribe("foo")
		_, subscription2, _ := pubsub.Subscribe("bar")
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
		_, subscription1, _ := pubsub.Subscribe("foo")

		pubsub.Publish("foo", "world")
		message1 := <-subscription1
		if message1 != "world" {
			t.Errorf("got message %v, want world", message1)
		}
	})

	t.Run("susbcriptions receive messages in order", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		_, subscription1, _ := pubsub.Subscribe("foo")
		pubsub.Publish("foo", "hello")
		pubsub.Publish("foo", "darkness")
		pubsub.Publish("foo", "my")
		pubsub.Publish("foo", "old")
		pubsub.Publish("foo", "friend")

		expected := []string{"hello", "darkness", "my", "old", "friend"}
		for _, want := range expected {
			message := <-subscription1
			if message != want {
				t.Errorf("got message %v, want %v", message, want)
			}
		}
	})

	t.Run("unsubscribing", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		id, subscription1, _ := pubsub.Subscribe("foo")
		pubsub.Publish("foo", "hello")
		message1 := <-subscription1
		if message1 != "hello" {
			t.Errorf("got message %v, want hello", message1)
		}
		ticker := time.NewTicker(100 * time.Millisecond)
		pubsub.Unsubscribe(id)
		pubsub.Publish("foo", "world")
		select {
		case message1 = <-subscription1:
			t.Errorf("expected no message, got %v", message1)
		case <-ticker.C:
			// ok
		}
	})

	t.Run("unsubscribing prevents writing to closed channel", func(t *testing.T) {
		pubsub := NewMemoryPubSub()
		id, _, _ := pubsub.Subscribe("foo")
		go func() {
			for {
				pubsub.Publish("foo", "world")
				time.Sleep(time.Millisecond)
			}
		}()
		// Unsubscribe after 1 second
		go func() {
			<-time.After(1 * time.Second)
			pubsub.Unsubscribe(id)
		}()
		<-time.After(1500 * time.Millisecond)
	})
}
