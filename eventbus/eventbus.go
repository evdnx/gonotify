package eventbus

import (
	"sync"
	"time"
)

// EventType represents the type of a published event.
type EventType string

// Predefined event types supported by the notification service.
const (
	EventTradeExecuted  EventType = "trade_executed"
	EventOrderFilled    EventType = "order_filled"
	EventPositionOpened EventType = "position_opened"
	EventPositionClosed EventType = "position_closed"
	EventPnLUpdate      EventType = "pnl_update"
	EventSystemError    EventType = "system_error"
	EventStrategyError  EventType = "strategy_error"
)

// Event encapsulates a payload broadcast on the EventBus.
type Event struct {
	Type      EventType
	Data      interface{}
	Timestamp time.Time
}

// EventHandler handles a published event.
type EventHandler func(Event)

// EventBus is a minimal publish/subscribe message bus for notifications.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType]map[string]EventHandler
}

// NewEventBus constructs an EventBus with no subscribers.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType]map[string]EventHandler),
	}
}

// Subscribe registers a handler for the given event type under a subscriber ID.
func (b *EventBus) Subscribe(eventType EventType, subscriberID string, handler EventHandler) {
	if handler == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[eventType]; !ok {
		b.subscribers[eventType] = make(map[string]EventHandler)
	}

	b.subscribers[eventType][subscriberID] = handler
}

// Unsubscribe removes a handler for the given event type.
func (b *EventBus) Unsubscribe(eventType EventType, subscriberID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, ok := b.subscribers[eventType]; ok {
		delete(handlers, subscriberID)
		if len(handlers) == 0 {
			delete(b.subscribers, eventType)
		}
	}
}

// Publish broadcasts an event to the registered subscribers.
func (b *EventBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	b.mu.RLock()
	handlers := b.subscribers[event.Type]
	copied := make([]EventHandler, 0, len(handlers))
	for _, handler := range handlers {
		copied = append(copied, handler)
	}
	b.mu.RUnlock()

	for _, handler := range copied {
		handler(event)
	}
}

// PublishData is a helper that publishes an event with the provided data.
func (b *EventBus) PublishData(eventType EventType, data interface{}) {
	b.Publish(Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	})
}

