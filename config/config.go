package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConfigFile represents the structure of the notification configuration file
type ConfigFile struct {
	Element  *ElementConfig  `json:"element,omitempty"`
	Telegram *TelegramConfig `json:"telegram,omitempty"`
	Events   EventConfig     `json:"events"`
}

// ElementConfig contains Element messenger configuration
type ElementConfig struct {
	HomeserverURL string `json:"homeserver_url"`
	AccessToken   string `json:"access_token"`
	RoomID        string `json:"room_id"`
	Enabled       bool   `json:"enabled"`
}

// TelegramConfig contains Telegram messenger configuration
type TelegramConfig struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// EventConfig contains event notification configuration
type EventConfig struct {
	TradeExecution  bool    `json:"trade_execution"`
	OrderFilled     bool    `json:"order_filled"`
	PositionChange  bool    `json:"position_change"`
	PnLUpdate       bool    `json:"pnl_update"`
	StopLoss        bool    `json:"stop_loss"`
	TakeProfit      bool    `json:"take_profit"`
	SystemErrors    bool    `json:"system_errors"`
	StrategyErrors  bool    `json:"strategy_errors"`
	ProfitThreshold float64 `json:"profit_threshold"`
}

// NotificationConfig contains configuration for the notification service
type NotificationConfig struct {
	// Element messenger configuration
	ElementHomeserverURL string
	ElementAccessToken   string
	ElementRoomID        string
	ElementEnabled       bool

	// Telegram messenger configuration
	TelegramBotToken string
	TelegramChatID   string
	TelegramEnabled  bool

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
		ElementRoomID:        "!cryptobot:matrix.org",
		ElementEnabled:       false,

		TelegramEnabled: false,

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

// LoadConfig loads notification configuration from a file
func LoadConfig(filePath string) (*NotificationConfig, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("notification config file not found: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notification config file: %w", err)
	}

	// Parse JSON
	var configFile ConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("failed to parse notification config file: %w", err)
	}

	// Convert to NotificationConfig
	config := &NotificationConfig{
		NotifyTradeExecution: configFile.Events.TradeExecution,
		NotifyOrderFilled:    configFile.Events.OrderFilled,
		NotifyPositionChange: configFile.Events.PositionChange,
		NotifyPnLUpdate:      configFile.Events.PnLUpdate,
		NotifyStopLoss:       configFile.Events.StopLoss,
		NotifyTakeProfit:     configFile.Events.TakeProfit,
		NotifySystemErrors:   configFile.Events.SystemErrors,
		NotifyStrategyErrors: configFile.Events.StrategyErrors,
		ProfitThreshold:      configFile.Events.ProfitThreshold,
	}

	// Load Element config if present
	if configFile.Element != nil {
		config.ElementHomeserverURL = configFile.Element.HomeserverURL
		config.ElementAccessToken = configFile.Element.AccessToken
		config.ElementRoomID = configFile.Element.RoomID
		config.ElementEnabled = configFile.Element.Enabled
	}

	// Load Telegram config if present
	if configFile.Telegram != nil {
		config.TelegramBotToken = configFile.Telegram.BotToken
		config.TelegramChatID = configFile.Telegram.ChatID
		config.TelegramEnabled = configFile.Telegram.Enabled
	}

	return config, nil
}

// SaveConfig saves notification configuration to a file
func SaveConfig(config *NotificationConfig, filePath string) error {
	// Convert to ConfigFile
	configFile := ConfigFile{
		Events: EventConfig{
			TradeExecution:  config.NotifyTradeExecution,
			OrderFilled:     config.NotifyOrderFilled,
			PositionChange:  config.NotifyPositionChange,
			PnLUpdate:       config.NotifyPnLUpdate,
			StopLoss:        config.NotifyStopLoss,
			TakeProfit:      config.NotifyTakeProfit,
			SystemErrors:    config.NotifySystemErrors,
			StrategyErrors:  config.NotifyStrategyErrors,
			ProfitThreshold: config.ProfitThreshold,
		},
	}

	// Add Element config if enabled
	if config.ElementEnabled {
		configFile.Element = &ElementConfig{
			HomeserverURL: config.ElementHomeserverURL,
			AccessToken:   config.ElementAccessToken,
			RoomID:        config.ElementRoomID,
			Enabled:       config.ElementEnabled,
		}
	}

	// Add Telegram config if enabled
	if config.TelegramEnabled {
		configFile.Telegram = &TelegramConfig{
			BotToken: config.TelegramBotToken,
			ChatID:   config.TelegramChatID,
			Enabled:  config.TelegramEnabled,
		}
	}

	// Convert to JSON
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notification config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write notification config file: %w", err)
	}

	return nil
}

// CreateDefaultConfigFile creates a default notification configuration file
func CreateDefaultConfigFile(filePath string) error {
	config := DefaultNotificationConfig()
	return SaveConfig(config, filePath)
}

