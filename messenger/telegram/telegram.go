package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is a client for sending messages to Telegram
type Client struct {
	botToken   string
	chatID     string
	httpClient *http.Client
	apiURL     string
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
		chatID:   chatID,
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

// SendFile sends a file to the Telegram chat using sendDocument API
func (c *Client) SendFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add chat_id field
	err = writer.WriteField("chat_id", c.chatID)
	if err != nil {
		return fmt.Errorf("failed to write chat_id field: %w", err)
	}

	// Add document field
	filename := filepath.Base(filePath)
	part, err := writer.CreateFormFile("document", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content to the form
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close the writer to finalize the multipart message
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Create the HTTP request
	url := fmt.Sprintf("%s/bot%s/sendDocument", c.apiURL, c.botToken)
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set content type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API error: status %d, body: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

// Name returns the name of the messenger
func (c *Client) Name() string {
	return "Telegram"
}
