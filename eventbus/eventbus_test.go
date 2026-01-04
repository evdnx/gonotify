package eventbus

import (
	"sync"
	"testing"
)

func TestEventBusPublishSubscribe(t *testing.T) {
	bus := NewEventBus()

	wg := sync.WaitGroup{}
	wg.Add(1)

	var received Event
	bus.Subscribe(EventTradeExecuted, "test-handler", func(e Event) {
		received = e
		wg.Done()
	})

	bus.PublishData(EventTradeExecuted, "payload")
	wg.Wait()

	if received.Type != EventTradeExecuted {
		t.Fatalf("expected event type %s, got %s", EventTradeExecuted, received.Type)
	}

	payload, ok := received.Data.(string)
	if !ok || payload != "payload" {
		t.Fatalf("unexpected payload: %#v", received.Data)
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	bus := NewEventBus()

	called := false
	bus.Subscribe(EventPnLUpdate, "listener", func(e Event) {
		called = true
	})

	bus.Unsubscribe(EventPnLUpdate, "listener")
	bus.PublishData(EventPnLUpdate, nil)

	if called {
		t.Fatal("handler should not have been invoked after unsubscribe")
	}
}

