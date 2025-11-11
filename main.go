package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ElementMessenger defines the contract used by NotificationService to send messages.
type ElementMessenger interface {
	SendMessage(message string) error
}

// ElementClient is a client for sending messages to Element messenger
type ElementClient struct {
	homeserverURL string
	accessToken   string
	roomID        string
	httpClient    *http.Client
}

// ElementMessage represents a message to be sent to Element
type ElementMessage struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// NewElementClient creates a new Element client
func NewElementClient(homeserverURL, accessToken, roomID string) *ElementClient {
	return &ElementClient{
		homeserverURL: homeserverURL,
		accessToken:   accessToken,
		roomID:        roomID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMessage sends a message to the Element chat room
func (c *ElementClient) SendMessage(message string) error {
	// Create the message payload
	payload := ElementMessage{
		MsgType: "m.text",
		Body:    message,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal message payload: %w", err)
	}

	// Create the request URL
	// Format: /_matrix/client/r0/rooms/{roomId}/send/m.room.message/{txnId}
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	url := fmt.Sprintf("%s/_matrix/client/r0/rooms/%s/send/m.room.message/%s?access_token=%s",
		c.homeserverURL, c.roomID, txnID, c.accessToken)

	// Create the request
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonPayload))
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

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}
