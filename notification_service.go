package notifications

import (
	"fmt"
	"time"
)

// NotificationService handles sending notifications for important events
type NotificationService struct {
	elementClient *ElementClient
	eventBus      *EventBus
	config        *NotificationConfig
}

// NotificationConfig contains configuration for the notification service
type NotificationConfig struct {
	// Element messenger configuration
	ElementHomeserverURL string
	ElementAccessToken   string
	ElementRoomID        string

	// Event types to notify about
	NotifyTradeExecution bool
	NotifyOrderFilled    bool
	NotifyPositionChange bool
	NotifyPnLUpdate      bool
	NotifyStopLoss       bool
	NotifyTakeProfit     bool
	NotifySystemErrors   bool
	NotifyStrategyErrors bool

	// Minimum profit threshold for PnL notifications (as a percentage)
	ProfitThreshold float64
}

// DefaultNotificationConfig returns a default notification configuration
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		ElementHomeserverURL: "https://matrix.org",
		ElementRoomID:        "!cryptobot:matrix.org", // Example room ID; replace with your own

		NotifyTradeExecution: true,
		NotifyOrderFilled:    true,
		NotifyPositionChange: true,
		NotifyPnLUpdate:      true,
		NotifyStopLoss:       true,
		NotifyTakeProfit:     true,
		NotifySystemErrors:   true,
		NotifyStrategyErrors: true,

		ProfitThreshold: 1.0, // 1% profit threshold
	}
}

// NewNotificationService creates a new notification service
func NewNotificationService(config *NotificationConfig, eventBus *EventBus) (*NotificationService, error) {
	if config == nil {
		config = DefaultNotificationConfig()
	}

	// Validate configuration
	if config.ElementHomeserverURL == "" {
		return nil, fmt.Errorf("element homeserver URL is required")
	}
	if config.ElementAccessToken == "" {
		return nil, fmt.Errorf("element access token is required")
	}
	if config.ElementRoomID == "" {
		return nil, fmt.Errorf("element room ID is required")
	}

	if eventBus == nil {
		eventBus = NewEventBus()
	}

	// Create Element client
	elementClient := NewElementClient(
		config.ElementHomeserverURL,
		config.ElementAccessToken,
		config.ElementRoomID,
	)

	return &NotificationService{
		elementClient: elementClient,
		eventBus:      eventBus,
		config:        config,
	}, nil
}

// Start registers event handlers and starts the notification service
func (s *NotificationService) Start() error {
	// Send a startup notification
	err := s.elementClient.SendMessage("🤖 Notification service started")
	if err != nil {
		return fmt.Errorf("failed to send startup notification: %w", err)
	}

	// Register event handlers
	s.registerEventHandlers()

	return nil
}

// registerEventHandlers registers handlers for events that should trigger notifications
func (s *NotificationService) registerEventHandlers() {
	// Create a filtered handler for trade events
	if s.config.NotifyTradeExecution {
		s.eventBus.Subscribe(EventTradeExecuted, "notification_service", s.handleTradeExecuted)
	}

	// Create a filtered handler for order events
	if s.config.NotifyOrderFilled {
		s.eventBus.Subscribe(EventOrderFilled, "notification_service", s.handleOrderFilled)
	}

	// Create a filtered handler for position events
	if s.config.NotifyPositionChange {
		s.eventBus.Subscribe(EventPositionOpened, "notification_service", s.handlePositionOpened)
		s.eventBus.Subscribe(EventPositionClosed, "notification_service", s.handlePositionClosed)
	}

	// Create a filtered handler for PnL events
	if s.config.NotifyPnLUpdate {
		s.eventBus.Subscribe(EventPnLUpdate, "notification_service", s.handlePnLUpdate)
	}

	// Create a filtered handler for system errors
	if s.config.NotifySystemErrors {
		s.eventBus.Subscribe(EventSystemError, "notification_service", s.handleSystemError)
	}

	// Create a filtered handler for strategy errors
	if s.config.NotifyStrategyErrors {
		s.eventBus.Subscribe(EventStrategyError, "notification_service", s.handleStrategyError)
	}
}

// handleTradeExecuted handles trade executed events
func (s *NotificationService) handleTradeExecuted(event Event) {
	// Extract trade data from the event
	var trade Trade

	// Try to convert the event data to our local Trade type
	if tradeData, ok := event.Data.(map[string]interface{}); ok {
		// Extract fields from the map
		if id, ok := tradeData["id"].(string); ok {
			trade.ID = id
		}
		if symbol, ok := tradeData["symbol"].(string); ok {
			trade.Symbol = symbol
		}
		if side, ok := tradeData["side"].(string); ok {
			trade.Side = side
		}
		if price, ok := tradeData["price"].(float64); ok {
			trade.Price = price
		}
		if quantity, ok := tradeData["quantity"].(float64); ok {
			trade.Quantity = quantity
		}
		if baseAsset, ok := tradeData["base_asset"].(string); ok {
			trade.BaseAsset = baseAsset
		}
		if quoteAsset, ok := tradeData["quote_asset"].(string); ok {
			trade.QuoteAsset = quoteAsset
		}
	} else {
		s.sendNotification("⚠️ Received malformed trade execution event")
		return
	}

	// Format the notification message
	message := fmt.Sprintf("💰 Trade Executed: %s %s %.6f %s at price %.2f %s",
		trade.Side, trade.Symbol, trade.Quantity, trade.BaseAsset, trade.Price, trade.QuoteAsset)

	// Send the notification
	s.sendNotification(message)
}

// handleOrderFilled handles order filled events
func (s *NotificationService) handleOrderFilled(event Event) {
	// Extract order data from the event
	var order Order

	// Try to convert the event data to our local Order type
	if orderData, ok := event.Data.(map[string]interface{}); ok {
		// Extract fields from the map
		if id, ok := orderData["id"].(string); ok {
			order.ID = id
		}
		if symbol, ok := orderData["symbol"].(string); ok {
			order.Symbol = symbol
		}
		if side, ok := orderData["side"].(string); ok {
			order.Side = side
		}
		if orderType, ok := orderData["type"].(string); ok {
			order.Type = orderType
		}
		if quantity, ok := orderData["quantity"].(float64); ok {
			order.Quantity = quantity
		}
		if price, ok := orderData["price"].(float64); ok {
			order.Price = price
		}
		if executedPrice, ok := orderData["executed_price"].(float64); ok {
			order.ExecutedPrice = executedPrice
		}
	} else {
		s.sendNotification("⚠️ Received malformed order filled event")
		return
	}

	// Check if this is a stop loss or take profit order
	isStopLoss := order.Type == "stop" || order.Type == "stop_market"
	isTakeProfit := order.Type == "take_profit" || order.Type == "take_profit_market"

	// Skip if we don't want to notify about this type
	if isStopLoss && !s.config.NotifyStopLoss {
		return
	}
	if isTakeProfit && !s.config.NotifyTakeProfit {
		return
	}

	// Format the notification message
	var emoji string
	if isStopLoss {
		emoji = "🛑"
	} else if isTakeProfit {
		emoji = "🎯"
	} else {
		emoji = "📝"
	}

	message := fmt.Sprintf("%s Order Filled: %s %s %.6f at price %.2f",
		emoji, order.Side, order.Symbol, order.Quantity, order.ExecutedPrice)

	// Send the notification
	s.sendNotification(message)
}

// handlePositionOpened handles position opened events
func (s *NotificationService) handlePositionOpened(event Event) {
	// Extract position data from the event
	var position Position

	// Try to convert the event data to our local Position type
	if posData, ok := event.Data.(map[string]interface{}); ok {
		// Extract fields from the map
		if id, ok := posData["id"].(string); ok {
			position.ID = id
		}
		if symbol, ok := posData["symbol"].(string); ok {
			position.Symbol = symbol
		}
		if side, ok := posData["side"].(string); ok {
			position.Side = side
		}
		if quantity, ok := posData["quantity"].(float64); ok {
			position.Quantity = quantity
		}
		if entryPrice, ok := posData["entry_price"].(float64); ok {
			position.EntryPrice = entryPrice
		}
	} else {
		s.sendNotification("⚠️ Received malformed position opened event")
		return
	}

	// Format the notification message
	message := fmt.Sprintf("🔓 Position Opened: %s %s %.6f at entry price %.2f",
		position.Side, position.Symbol, position.Quantity, position.EntryPrice)

	// Send the notification
	s.sendNotification(message)
}

// handlePositionClosed handles position closed events
func (s *NotificationService) handlePositionClosed(event Event) {
	// Extract position data from the event
	var position Position

	// Try to convert the event data to our local Position type
	if posData, ok := event.Data.(map[string]interface{}); ok {
		// Extract fields from the map
		if id, ok := posData["id"].(string); ok {
			position.ID = id
		}
		if symbol, ok := posData["symbol"].(string); ok {
			position.Symbol = symbol
		}
		if side, ok := posData["side"].(string); ok {
			position.Side = side
		}
		if quantity, ok := posData["quantity"].(float64); ok {
			position.Quantity = quantity
		}
		if entryPrice, ok := posData["entry_price"].(float64); ok {
			position.EntryPrice = entryPrice
		}
		if exitPrice, ok := posData["exit_price"].(float64); ok {
			position.ExitPrice = exitPrice
		}
		if realizedPnL, ok := posData["realized_pnl"].(float64); ok {
			position.RealizedPnL = realizedPnL
		}
	} else {
		s.sendNotification("⚠️ Received malformed position closed event")
		return
	}

	// Calculate profit/loss
	pnl := position.RealizedPnL
	pnlPercentage := (position.ExitPrice - position.EntryPrice) / position.EntryPrice * 100
	if position.Side == "sell" {
		pnlPercentage = -pnlPercentage
	}

	// Format the notification message
	var emoji string
	if pnl > 0 {
		emoji = "🔒💰"
	} else {
		emoji = "🔒📉"
	}

	message := fmt.Sprintf("%s Position Closed: %s %s %.6f at exit price %.2f (P&L: %.2f / %.2f%%)",
		emoji, position.Side, position.Symbol, position.Quantity, position.ExitPrice, pnl, pnlPercentage)

	// Send the notification
	s.sendNotification(message)
}

// handlePnLUpdate handles PnL update events
func (s *NotificationService) handlePnLUpdate(event Event) {
	// Extract PnL data from the event
	pnlData, ok := event.Data.(map[string]interface{})
	if !ok {
		s.sendNotification("⚠️ Received malformed PnL update event")
		return
	}

	// Extract relevant fields
	symbol, _ := pnlData["symbol"].(string)
	pnl, _ := pnlData["pnl"].(float64)
	pnlPercentage, _ := pnlData["pnl_percentage"].(float64)

	// Only notify if profit/loss exceeds threshold
	if pnlPercentage < s.config.ProfitThreshold && pnlPercentage > -s.config.ProfitThreshold {
		return
	}

	// Format the notification message
	var emoji string
	if pnl > 0 {
		emoji = "📈"
	} else {
		emoji = "📉"
	}

	message := fmt.Sprintf("%s P&L Update for %s: %.2f (%.2f%%)",
		emoji, symbol, pnl, pnlPercentage)

	// Send the notification
	s.sendNotification(message)
}

// handleSystemError handles system error events
func (s *NotificationService) handleSystemError(event Event) {
	// Extract error data from the event
	errorMsg, ok := event.Data.(string)
	if !ok {
		s.sendNotification("⚠️ Received malformed system error event")
		return
	}

	// Format the notification message
	message := fmt.Sprintf("🚨 System Error: %s", errorMsg)

	// Send the notification
	s.sendNotification(message)
}

// handleStrategyError handles strategy error events
func (s *NotificationService) handleStrategyError(event Event) {
	// Extract error data from the event
	errorData, ok := event.Data.(map[string]interface{})
	if !ok {
		s.sendNotification("⚠️ Received malformed strategy error event")
		return
	}

	// Extract relevant fields
	strategy, _ := errorData["strategy"].(string)
	errorMsg, _ := errorData["error"].(string)

	// Format the notification message
	message := fmt.Sprintf("🚨 Strategy Error in %s: %s", strategy, errorMsg)

	// Send the notification
	s.sendNotification(message)
}

// sendNotification sends a notification message to Element
func (s *NotificationService) sendNotification(message string) {
	// Add timestamp to the message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Send the message asynchronously
	go func() {
		err := s.elementClient.SendMessage(fullMessage)
		if err != nil {
			// Log the error but don't propagate it
			fmt.Printf("Failed to send notification: %v\n", err)
		}
	}()
}
