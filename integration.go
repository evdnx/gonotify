package gonotify

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/evdnx/gonotify/config"
	"github.com/evdnx/gonotify/eventbus"
	"github.com/evdnx/gonotify/service"
)

// InitializeNotificationSystem initializes and starts the notification system
// This function is meant to be called during application startup
// If telegramBotToken and telegramChatID are provided and telegram is not already configured,
// they will be used to enable telegram notifications.
func InitializeNotificationSystem(eventBus *eventbus.EventBus, configPath string, telegramBotToken, telegramChatID string) (*service.NotificationService, error) {
	if eventBus == nil {
		eventBus = eventbus.NewEventBus()
	}

	// If configPath is empty, use default path
	if configPath == "" {
		// Try to find the config file in common locations
		locations := []string{
			"configs/notification.json",
			"../configs/notification.json",
			"../../configs/notification.json",
		}

		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				configPath = loc
				break
			}
		}

		// If still not found, use the first option as default
		if configPath == "" {
			configPath = locations[0]
		}
	}

	// Ensure the config directory exists
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Check if config file exists, create default if not
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Notification config file not found, creating default at %s\n", configPath)
		if err := config.CreateDefaultConfigFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default notification config created at %s\n", configPath)
		fmt.Println("Please update the config file with your messenger credentials")
	}

	// Load notification configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load notification config: %w", err)
	}

	// Check if access tokens are provided via environment variables
	if token := os.Getenv("ELEMENT_ACCESS_TOKEN"); token != "" {
		cfg.ElementAccessToken = token
		if cfg.ElementHomeserverURL != "" && cfg.ElementRoomID != "" {
			cfg.ElementEnabled = true
		}
	}

	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.TelegramBotToken = token
		if chatID := os.Getenv("TELEGRAM_CHAT_ID"); chatID != "" {
			cfg.TelegramChatID = chatID
			cfg.TelegramEnabled = true
		}
	}

	// If telegram is not already enabled/configured, use function parameters if provided
	if !cfg.TelegramEnabled && telegramBotToken != "" && telegramChatID != "" {
		cfg.TelegramBotToken = telegramBotToken
		cfg.TelegramChatID = telegramChatID
		cfg.TelegramEnabled = true
	}

	// Validate at least one messenger is configured
	if !cfg.ElementEnabled && !cfg.TelegramEnabled {
		return nil, fmt.Errorf("at least one messenger must be enabled. Please update %s or set environment variables", configPath)
	}

	// Validate Element config if enabled
	if cfg.ElementEnabled {
		if cfg.ElementAccessToken == "" || cfg.ElementAccessToken == "YOUR_ELEMENT_ACCESS_TOKEN" {
			return nil, fmt.Errorf("Element access token not provided. Please update %s or set ELEMENT_ACCESS_TOKEN environment variable", configPath)
		}
		if cfg.ElementHomeserverURL == "" {
			return nil, fmt.Errorf("Element homeserver URL not provided in %s", configPath)
		}
		if cfg.ElementRoomID == "" {
			return nil, fmt.Errorf("Element room ID not provided in %s", configPath)
		}
	}

	// Validate Telegram config if enabled
	if cfg.TelegramEnabled {
		if cfg.TelegramBotToken == "" || cfg.TelegramBotToken == "YOUR_TELEGRAM_BOT_TOKEN" {
			return nil, fmt.Errorf("Telegram bot token not provided. Please update %s or set TELEGRAM_BOT_TOKEN environment variable", configPath)
		}
		if cfg.TelegramChatID == "" {
			return nil, fmt.Errorf("Telegram chat ID not provided. Please update %s or set TELEGRAM_CHAT_ID environment variable", configPath)
		}
	}

	// Create notification service
	notificationService, err := service.NewNotificationService(cfg, eventBus)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification service: %w", err)
	}

	// Start notification service
	if err := notificationService.Start(); err != nil {
		return nil, fmt.Errorf("failed to start notification service: %w", err)
	}

	fmt.Println("Notification system initialized successfully")
	return notificationService, nil
}
