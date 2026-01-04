package element

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a client for sending messages to Element (Matrix) messenger
type Client struct {
	homeserverURL string
	accessToken   string
	roomID        string
	httpClient    *http.Client
}

// Message represents a message to be sent to Element
type Message struct {
	MsgType string `json:"msgtype"`
	Body    string `json:"body"`
}

// NewClient creates a new Element client
func NewClient(homeserverURL, accessToken, roomID string) *Client {
	return &Client{
		homeserverURL: homeserverURL,
		accessToken:   accessToken,
		roomID:        roomID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendMessage sends a message to the Element chat room
func (c *Client) SendMessage(message string) error {
	// Create the message payload
	payload := Message{
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

// Name returns the name of the messenger
func (c *Client) Name() string {
	return "Element"
}

