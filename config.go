package notifications

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConfigFile represents the structure of the notification configuration file
type ConfigFile struct {
	Element ElementConfig `json:"element"`
	Events  EventConfig   `json:"events"`
}

// ElementConfig contains Element messenger configuration
type ElementConfig struct {
	HomeserverURL string `json:"homeserver_url"`
	AccessToken   string `json:"access_token"`
	RoomID        string `json:"room_id"`
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
		ElementHomeserverURL: configFile.Element.HomeserverURL,
		ElementAccessToken:   configFile.Element.AccessToken,
		ElementRoomID:        configFile.Element.RoomID,

		NotifyTradeExecution: configFile.Events.TradeExecution,
		NotifyOrderFilled:    configFile.Events.OrderFilled,
		NotifyPositionChange: configFile.Events.PositionChange,
		NotifyPnLUpdate:      configFile.Events.PnLUpdate,
		NotifyStopLoss:       configFile.Events.StopLoss,
		NotifyTakeProfit:     configFile.Events.TakeProfit,
		NotifySystemErrors:   configFile.Events.SystemErrors,
		NotifyStrategyErrors: configFile.Events.StrategyErrors,

		ProfitThreshold: configFile.Events.ProfitThreshold,
	}

	return config, nil
}

// SaveConfig saves notification configuration to a file
func SaveConfig(config *NotificationConfig, filePath string) error {
	// Convert to ConfigFile
	configFile := ConfigFile{
		Element: ElementConfig{
			HomeserverURL: config.ElementHomeserverURL,
			AccessToken:   config.ElementAccessToken,
			RoomID:        config.ElementRoomID,
		},
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
