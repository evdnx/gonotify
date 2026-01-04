package config

import (
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "notification.json")

	original := &NotificationConfig{
		ElementHomeserverURL: "https://matrix.org",
		ElementAccessToken:   "token",
		ElementRoomID:        "!room:id",
		ElementEnabled:       true,
		TelegramBotToken:     "bot_token",
		TelegramChatID:       "chat_id",
		TelegramEnabled:      true,
		NotifyTradeExecution: true,
		NotifyOrderFilled:    true,
		NotifyPositionChange: false,
		NotifyPnLUpdate:      true,
		NotifyStopLoss:       false,
		NotifyTakeProfit:     true,
		NotifySystemErrors:   true,
		NotifyStrategyErrors: false,
		ProfitThreshold:      2.5,
	}

	if err := SaveConfig(original, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.ElementHomeserverURL != original.ElementHomeserverURL ||
		loaded.ElementAccessToken != original.ElementAccessToken ||
		loaded.ElementRoomID != original.ElementRoomID ||
		loaded.ElementEnabled != original.ElementEnabled ||
		loaded.TelegramBotToken != original.TelegramBotToken ||
		loaded.TelegramChatID != original.TelegramChatID ||
		loaded.TelegramEnabled != original.TelegramEnabled ||
		loaded.NotifyTradeExecution != original.NotifyTradeExecution ||
		loaded.NotifyOrderFilled != original.NotifyOrderFilled ||
		loaded.NotifyPositionChange != original.NotifyPositionChange ||
		loaded.NotifyPnLUpdate != original.NotifyPnLUpdate ||
		loaded.NotifyStopLoss != original.NotifyStopLoss ||
		loaded.NotifyTakeProfit != original.NotifyTakeProfit ||
		loaded.NotifySystemErrors != original.NotifySystemErrors ||
		loaded.NotifyStrategyErrors != original.NotifyStrategyErrors ||
		loaded.ProfitThreshold != original.ProfitThreshold {
		t.Fatal("loaded config does not match original")
	}
}

