package notifications

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitializeNotificationSystem initializes and starts the notification system
// This function is meant to be called during application startup
func InitializeNotificationSystem(eventBus *EventBus, configPath string) (*NotificationService, error) {
	if eventBus == nil {
		eventBus = NewEventBus()
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
		if err := CreateDefaultConfigFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config file: %w", err)
		}
		fmt.Printf("Default notification config created at %s\n", configPath)
		fmt.Println("Please update the config file with your Element access token and room ID")
	}

	// Load notification configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load notification config: %w", err)
	}

	// Check if access token is provided via environment variable
	if token := os.Getenv("ELEMENT_ACCESS_TOKEN"); token != "" {
		config.ElementAccessToken = token
	}

	// Validate access token
	if config.ElementAccessToken == "" || config.ElementAccessToken == "YOUR_ELEMENT_ACCESS_TOKEN" {
		return nil, fmt.Errorf("Element access token not provided. Please update %s or set ELEMENT_ACCESS_TOKEN environment variable", configPath)
	}

	// Create notification service
	notificationService, err := NewNotificationService(config, eventBus)
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
