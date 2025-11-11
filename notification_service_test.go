package notifications

import (
	"strings"
	"testing"
	"time"
)

type mockMessenger struct {
	ch chan string
}

func newMockMessenger() *mockMessenger {
	return &mockMessenger{
		ch: make(chan string, 10),
	}
}

func (m *mockMessenger) SendMessage(message string) error {
	m.ch <- message
	return nil
}

func (m *mockMessenger) waitForMessage(t *testing.T, contains string) string {
	t.Helper()

	select {
	case msg := <-m.ch:
		if !strings.Contains(msg, contains) {
			t.Fatalf("message %q does not contain %q", msg, contains)
		}
		return msg
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for message containing %q", contains)
	}
	return ""
}

func TestNotificationServiceSendsTradeNotification(t *testing.T) {
	config := &NotificationConfig{
		ElementHomeserverURL: "https://matrix.org",
		ElementAccessToken:   "token",
		ElementRoomID:        "!room:id",
		NotifyTradeExecution: true,
	}

	eventBus := NewEventBus()
	messenger := newMockMessenger()

	service, err := NewNotificationServiceWithMessenger(config, eventBus, messenger)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	if err := service.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	messenger.waitForMessage(t, "Notification service started")

	eventBus.PublishData(EventTradeExecuted, map[string]interface{}{
		"id":          "trade-1",
		"symbol":      "BTCUSDT",
		"side":        "buy",
		"quantity":    0.5,
		"price":       68000.0,
		"base_asset":  "BTC",
		"quote_asset": "USDT",
	})

	messenger.waitForMessage(t, "Trade Executed")
}

func TestNewNotificationServiceWithMessengerRejectsNil(t *testing.T) {
	_, err := NewNotificationServiceWithMessenger(DefaultNotificationConfig(), NewEventBus(), nil)
	if err == nil {
		t.Fatal("expected error when messenger is nil")
	}
}
