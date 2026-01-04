package service

import (
	"fmt"
	"time"

	"github.com/evdnx/gonotify/config"
	"github.com/evdnx/gonotify/eventbus"
	"github.com/evdnx/gonotify/messenger"
	"github.com/evdnx/gonotify/messenger/element"
	"github.com/evdnx/gonotify/messenger/telegram"
	"github.com/evdnx/gonotify/types"
)

// NotificationService handles sending notifications for important events
type NotificationService struct {
	messengers []messenger.Messenger
	eventBus   *eventbus.EventBus
	config     *config.NotificationConfig
}

// NewNotificationService creates a new notification service with messengers based on config.
func NewNotificationService(cfg *config.NotificationConfig, bus *eventbus.EventBus) (*NotificationService, error) {
	return newNotificationService(cfg, bus, nil)
}

// NewNotificationServiceWithMessengers creates a notification service using custom messenger implementations.
func NewNotificationServiceWithMessengers(cfg *config.NotificationConfig, bus *eventbus.EventBus, messengers []messenger.Messenger) (*NotificationService, error) {
	if len(messengers) == 0 {
		return nil, fmt.Errorf("at least one messenger is required")
	}
	return newNotificationService(cfg, bus, messengers)
}

func newNotificationService(cfg *config.NotificationConfig, bus *eventbus.EventBus, messengers []messenger.Messenger) (*NotificationService, error) {
	if cfg == nil {
		cfg = config.DefaultNotificationConfig()
	}

	if bus == nil {
		bus = eventbus.NewEventBus()
	}

	// Create messengers if not provided
	if messengers == nil {
		messengers = []messenger.Messenger{}

		// Create Element messenger if enabled
		if cfg.ElementEnabled {
			if cfg.ElementHomeserverURL == "" {
				return nil, fmt.Errorf("element homeserver URL is required when element is enabled")
			}
			if cfg.ElementAccessToken == "" {
				return nil, fmt.Errorf("element access token is required when element is enabled")
			}
			if cfg.ElementRoomID == "" {
				return nil, fmt.Errorf("element room ID is required when element is enabled")
			}
			messengers = append(messengers, element.NewClient(
				cfg.ElementHomeserverURL,
				cfg.ElementAccessToken,
				cfg.ElementRoomID,
			))
		}

		// Create Telegram messenger if enabled
		if cfg.TelegramEnabled {
			if cfg.TelegramBotToken == "" {
				return nil, fmt.Errorf("telegram bot token is required when telegram is enabled")
			}
			if cfg.TelegramChatID == "" {
				return nil, fmt.Errorf("telegram chat ID is required when telegram is enabled")
			}
			messengers = append(messengers, telegram.NewClient(
				cfg.TelegramBotToken,
				cfg.TelegramChatID,
			))
		}

		if len(messengers) == 0 {
			return nil, fmt.Errorf("at least one messenger must be enabled in configuration")
		}
	}

	return &NotificationService{
		messengers: messengers,
		eventBus:   bus,
		config:     cfg,
	}, nil
}

// Start registers event handlers and starts the notification service
func (s *NotificationService) Start() error {
	// Send a startup notification
	startupMsg := "🤖 Notification service started"
	s.sendNotification(startupMsg)

	// Register event handlers
	s.registerEventHandlers()

	return nil
}

// registerEventHandlers registers handlers for events that should trigger notifications
func (s *NotificationService) registerEventHandlers() {
	// Create a filtered handler for trade events
	if s.config.NotifyTradeExecution {
		s.eventBus.Subscribe(eventbus.EventTradeExecuted, "notification_service", s.handleTradeExecuted)
	}

	// Create a filtered handler for order events
	if s.config.NotifyOrderFilled {
		s.eventBus.Subscribe(eventbus.EventOrderFilled, "notification_service", s.handleOrderFilled)
	}

	// Create a filtered handler for position events
	if s.config.NotifyPositionChange {
		s.eventBus.Subscribe(eventbus.EventPositionOpened, "notification_service", s.handlePositionOpened)
		s.eventBus.Subscribe(eventbus.EventPositionClosed, "notification_service", s.handlePositionClosed)
	}

	// Create a filtered handler for PnL events
	if s.config.NotifyPnLUpdate {
		s.eventBus.Subscribe(eventbus.EventPnLUpdate, "notification_service", s.handlePnLUpdate)
	}

	// Create a filtered handler for system errors
	if s.config.NotifySystemErrors {
		s.eventBus.Subscribe(eventbus.EventSystemError, "notification_service", s.handleSystemError)
	}

	// Create a filtered handler for strategy errors
	if s.config.NotifyStrategyErrors {
		s.eventBus.Subscribe(eventbus.EventStrategyError, "notification_service", s.handleStrategyError)
	}
}

// handleTradeExecuted handles trade executed events
func (s *NotificationService) handleTradeExecuted(event eventbus.Event) {
	// Try to extract trade data
	var trade types.Trade
	if err := s.extractTrade(event.Data, &trade); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed trade execution event: %v", err))
		return
	}

	// Format the notification message
	message := fmt.Sprintf("💰 Trade Executed: %s %s %.6f %s at price %.2f %s",
		trade.Side, trade.Symbol, trade.Quantity, trade.BaseAsset, trade.Price, trade.QuoteAsset)

	// Send the notification
	s.sendNotification(message)
}

// handleOrderFilled handles order filled events
func (s *NotificationService) handleOrderFilled(event eventbus.Event) {
	// Try to extract order data
	var order types.Order
	if err := s.extractOrder(event.Data, &order); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed order filled event: %v", err))
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
func (s *NotificationService) handlePositionOpened(event eventbus.Event) {
	// Try to extract position data
	var position types.Position
	if err := s.extractPosition(event.Data, &position); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed position opened event: %v", err))
		return
	}

	// Format the notification message
	message := fmt.Sprintf("🔓 Position Opened: %s %s %.6f at entry price %.2f",
		position.Side, position.Symbol, position.Quantity, position.EntryPrice)

	// Send the notification
	s.sendNotification(message)
}

// handlePositionClosed handles position closed events
func (s *NotificationService) handlePositionClosed(event eventbus.Event) {
	// Try to extract position data
	var position types.Position
	if err := s.extractPosition(event.Data, &position); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed position closed event: %v", err))
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
func (s *NotificationService) handlePnLUpdate(event eventbus.Event) {
	// Try to extract PnL data
	var pnlUpdate types.PnLUpdate
	if err := s.extractPnLUpdate(event.Data, &pnlUpdate); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed PnL update event: %v", err))
		return
	}

	// Only notify if profit/loss exceeds threshold
	if pnlUpdate.PnLPercentage < s.config.ProfitThreshold && pnlUpdate.PnLPercentage > -s.config.ProfitThreshold {
		return
	}

	// Format the notification message
	var emoji string
	if pnlUpdate.PnL > 0 {
		emoji = "📈"
	} else {
		emoji = "📉"
	}

	message := fmt.Sprintf("%s P&L Update for %s: %.2f (%.2f%%)",
		emoji, pnlUpdate.Symbol, pnlUpdate.PnL, pnlUpdate.PnLPercentage)

	// Send the notification
	s.sendNotification(message)
}

// handleSystemError handles system error events
func (s *NotificationService) handleSystemError(event eventbus.Event) {
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
func (s *NotificationService) handleStrategyError(event eventbus.Event) {
	// Try to extract strategy error data
	var strategyError types.StrategyError
	if err := s.extractStrategyError(event.Data, &strategyError); err != nil {
		s.sendNotification(fmt.Sprintf("⚠️ Received malformed strategy error event: %v", err))
		return
	}

	// Format the notification message
	message := fmt.Sprintf("🚨 Strategy Error in %s: %s", strategyError.Strategy, strategyError.Error)

	// Send the notification
	s.sendNotification(message)
}

// sendNotification sends a notification message to all configured messengers
func (s *NotificationService) sendNotification(message string) {
	// Add timestamp to the message
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Send the message asynchronously to all messengers
	for _, msg := range s.messengers {
		go func(m messenger.Messenger) {
			if err := m.SendMessage(fullMessage); err != nil {
				// Log the error but don't propagate it
				fmt.Printf("Failed to send notification via %s: %v\n", m.Name(), err)
			}
		}(msg)
	}
}

// Helper functions to extract typed data from interface{}
func (s *NotificationService) extractTrade(data interface{}, trade *types.Trade) error {
	if tradeData, ok := data.(map[string]interface{}); ok {
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
		return nil
	}
	// Try direct type assertion
	if t, ok := data.(*types.Trade); ok {
		*trade = *t
		return nil
	}
	if t, ok := data.(types.Trade); ok {
		*trade = t
		return nil
	}
	return fmt.Errorf("cannot extract trade from data")
}

func (s *NotificationService) extractOrder(data interface{}, order *types.Order) error {
	if orderData, ok := data.(map[string]interface{}); ok {
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
		return nil
	}
	// Try direct type assertion
	if o, ok := data.(*types.Order); ok {
		*order = *o
		return nil
	}
	if o, ok := data.(types.Order); ok {
		*order = o
		return nil
	}
	return fmt.Errorf("cannot extract order from data")
}

func (s *NotificationService) extractPosition(data interface{}, position *types.Position) error {
	if posData, ok := data.(map[string]interface{}); ok {
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
		return nil
	}
	// Try direct type assertion
	if p, ok := data.(*types.Position); ok {
		*position = *p
		return nil
	}
	if p, ok := data.(types.Position); ok {
		*position = p
		return nil
	}
	return fmt.Errorf("cannot extract position from data")
}

func (s *NotificationService) extractPnLUpdate(data interface{}, pnlUpdate *types.PnLUpdate) error {
	if pnlData, ok := data.(map[string]interface{}); ok {
		if symbol, ok := pnlData["symbol"].(string); ok {
			pnlUpdate.Symbol = symbol
		}
		if pnl, ok := pnlData["pnl"].(float64); ok {
			pnlUpdate.PnL = pnl
		}
		if pnlPercentage, ok := pnlData["pnl_percentage"].(float64); ok {
			pnlUpdate.PnLPercentage = pnlPercentage
		}
		return nil
	}
	// Try direct type assertion
	if p, ok := data.(*types.PnLUpdate); ok {
		*pnlUpdate = *p
		return nil
	}
	if p, ok := data.(types.PnLUpdate); ok {
		*pnlUpdate = p
		return nil
	}
	return fmt.Errorf("cannot extract PnL update from data")
}

func (s *NotificationService) extractStrategyError(data interface{}, strategyError *types.StrategyError) error {
	if errorData, ok := data.(map[string]interface{}); ok {
		if strategy, ok := errorData["strategy"].(string); ok {
			strategyError.Strategy = strategy
		}
		if errorMsg, ok := errorData["error"].(string); ok {
			strategyError.Error = errorMsg
		}
		return nil
	}
	// Try direct type assertion
	if se, ok := data.(*types.StrategyError); ok {
		*strategyError = *se
		return nil
	}
	if se, ok := data.(types.StrategyError); ok {
		*strategyError = se
		return nil
	}
	return fmt.Errorf("cannot extract strategy error from data")
}

