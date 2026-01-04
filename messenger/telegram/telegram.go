package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a client for sending messages to Telegram
type Client struct {
	botToken  string
	chatID    string
	httpClient *http.Client
	apiURL    string
}

// Message represents a message to be sent to Telegram
type Message struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

// Response represents the response from Telegram API
type Response struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// NewClient creates a new Telegram client
func NewClient(botToken, chatID string) *Client {
	return &Client{
		botToken: botToken,
		chatID:  chatID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		apiURL: "https://api.telegram.org",
	}
}

// SendMessage sends a message to the Telegram chat
func (c *Client) SendMessage(message string) error {
	// Create the message payload
	payload := Message{
		ChatID: c.chatID,
		Text:   message,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	// Create the request URL
	url := fmt.Sprintf("%s/bot%s/sendMessage", c.apiURL, c.botToken)

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var tgResponse Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResponse); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check the response
	if resp.StatusCode != http.StatusOK || !tgResponse.OK {
		return fmt.Errorf("failed to send message: %s (status: %d)", tgResponse.Description, resp.StatusCode)
	}

	return nil
}

// Name returns the name of the messenger
func (c *Client) Name() string {
	return "Telegram"
}

