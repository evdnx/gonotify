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

func (m *mockMessenger) expectNoMessage(t *testing.T, duration time.Duration) {
	t.Helper()

	select {
	case msg := <-m.ch:
		t.Fatalf("unexpected message received: %q", msg)
	case <-time.After(duration):
	}
}

func testConfig() *NotificationConfig {
	return &NotificationConfig{
		ElementHomeserverURL: "https://matrix.org",
		ElementAccessToken:   "token",
		ElementRoomID:        "!room:id",
		NotifyTradeExecution: true,
		NotifyOrderFilled:    true,
		NotifyPositionChange: true,
		NotifyPnLUpdate:      true,
		NotifyStopLoss:       true,
		NotifyTakeProfit:     true,
		NotifySystemErrors:   true,
		NotifyStrategyErrors: true,
		ProfitThreshold:      1.0,
	}
}

func startTestService(t *testing.T, cfg *NotificationConfig) (*EventBus, *mockMessenger) {
	t.Helper()

	eventBus := NewEventBus()
	messenger := newMockMessenger()

	service, err := NewNotificationServiceWithMessenger(cfg, eventBus, messenger)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	if err := service.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	messenger.waitForMessage(t, "Notification service started")
	return eventBus, messenger
}

func TestNotificationServiceSendsTradeNotification(t *testing.T) {
	config := testConfig()
	eventBus, messenger := startTestService(t, config)

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

func TestOrderFilledNotification(t *testing.T) {
	config := testConfig()
	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventOrderFilled, map[string]interface{}{
		"id":             "order-1",
		"symbol":         "ETHUSD",
		"side":           "sell",
		"type":           "limit",
		"quantity":       1.0,
		"executed_price": 2000.0,
	})

	messenger.waitForMessage(t, "Order Filled")
}

func TestStopLossNotificationRespectsConfig(t *testing.T) {
	config := testConfig()
	config.NotifyStopLoss = false

	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventOrderFilled, map[string]interface{}{
		"id":             "order-stop",
		"symbol":         "BTCUSD",
		"side":           "sell",
		"type":           "stop",
		"quantity":       0.25,
		"executed_price": 30000.0,
	})

	messenger.expectNoMessage(t, 300*time.Millisecond)
}

func TestPositionOpenedNotification(t *testing.T) {
	config := testConfig()
	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventPositionOpened, map[string]interface{}{
		"id":          "pos-1",
		"symbol":      "SOLUSD",
		"side":        "buy",
		"quantity":    5.0,
		"entry_price": 150.0,
	})

	messenger.waitForMessage(t, "Position Opened")
}

func TestPositionClosedNotificationIncludesPnL(t *testing.T) {
	config := testConfig()
	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventPositionClosed, map[string]interface{}{
		"id":             "pos-1",
		"symbol":         "SOLUSD",
		"side":           "buy",
		"quantity":       5.0,
		"entry_price":    150.0,
		"exit_price":     165.0,
		"realized_pnl":   75.0,
		"unrealized_pnl": 0.0,
	})

	msg := messenger.waitForMessage(t, "Position Closed")
	if !strings.Contains(msg, "P&L:") {
		t.Fatalf("expected P&L details in message, got %q", msg)
	}
}

func TestPnLThreshold(t *testing.T) {
	config := testConfig()
	config.ProfitThreshold = 5.0

	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventPnLUpdate, map[string]interface{}{
		"symbol":         "ADAUSD",
		"pnl":            10.0,
		"pnl_percentage": 2.0,
	})
	messenger.expectNoMessage(t, 300*time.Millisecond)

	eventBus.PublishData(EventPnLUpdate, map[string]interface{}{
		"symbol":         "ADAUSD",
		"pnl":            30.0,
		"pnl_percentage": 6.0,
	})
	messenger.waitForMessage(t, "P&L Update")
}

func TestErrorNotifications(t *testing.T) {
	config := testConfig()
	eventBus, messenger := startTestService(t, config)

	eventBus.PublishData(EventSystemError, "connection lost")
	messenger.waitForMessage(t, "System Error")

	eventBus.PublishData(EventStrategyError, map[string]interface{}{
		"strategy": "mean-revert",
		"error":    "division by zero",
	})
	messenger.waitForMessage(t, "Strategy Error")
}
