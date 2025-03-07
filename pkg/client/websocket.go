package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType represents the type of message
type MessageType string

const (
	// ChatMessage is a chat message
	ChatMessage MessageType = "chat"
	// PingMessage is a ping message
	PingMessage MessageType = "ping"
	// CustomMessage is a custom message
	CustomMessage MessageType = "custom"
	// ScreenshotMessage is a screenshot message
	ScreenshotMessage MessageType = "screenshot"
)

// Message represents a message to be sent to the WebSocket server
type Message struct {
	Type           MessageType    `json:"type"`
	Message        string         `json:"message,omitempty"`
	Timestamp      string         `json:"timestamp,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	Platform       string         `json:"platform,omitempty"`
	Version        string         `json:"version,omitempty"`
	Extra          map[string]any `json:"extra,omitempty"`
	ScreenshotData string         `json:"screenshotData,omitempty"` // Base64-encoded screenshot data
	ImageFormat    string         `json:"imageFormat,omitempty"`    // Format of the image (e.g., "png", "jpeg")
	Width          int            `json:"width,omitempty"`          // Width of the screenshot
	Height         int            `json:"height,omitempty"`         // Height of the screenshot
}

// MessageHandler is a function that handles a specific type of message
type MessageHandler func(data []byte) error

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	URL            string
	Conn           *websocket.Conn
	Handlers       map[string]MessageHandler
	Connected      bool
	ConnectTimeout time.Duration
	Verbose        bool
	mu             sync.Mutex
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(url string, verbose bool) *WebSocketClient {
	return &WebSocketClient{
		URL:            url,
		Handlers:       make(map[string]MessageHandler),
		ConnectTimeout: 10 * time.Second,
		Verbose:        verbose,
	}
}

// Connect connects to the WebSocket server
func (c *WebSocketClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Connected {
		if c.Verbose {
			log.Printf("DEBUG: Already connected to WebSocket server at %s", c.URL)
		}
		return nil
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: c.ConnectTimeout,
	}

	if c.Verbose {
		log.Printf("DEBUG: Attempting to connect to WebSocket server at %s...", c.URL)
	}

	conn, resp, err := dialer.Dial(c.URL, nil)
	if err != nil {
		if resp != nil {
			log.Printf("ERROR: Failed to connect to WebSocket server. Status code: %d", resp.StatusCode)
		}
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	c.Conn = conn
	c.Connected = true

	if c.Verbose {
		log.Printf("DEBUG: Successfully connected to WebSocket server at %s", c.URL)
		log.Printf("DEBUG: Connection details: Local: %s, Remote: %s",
			conn.LocalAddr().String(), conn.RemoteAddr().String())
	}

	// Start message handler
	go c.handleMessages()

	return nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Connected || c.Conn == nil {
		return nil
	}

	err := c.Conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		return fmt.Errorf("error sending close message: %w", err)
	}

	err = c.Conn.Close()
	if err != nil {
		return fmt.Errorf("error closing connection: %w", err)
	}

	c.Connected = false
	return nil
}

// IsConnected returns whether the client is connected
func (c *WebSocketClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Connected
}

// RegisterHandler registers a handler for a specific message type
func (c *WebSocketClient) RegisterHandler(messageType string, handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Handlers[messageType] = handler
}

// SendJSON sends a JSON message to the server
func (c *WebSocketClient) SendJSON(message interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Connected || c.Conn == nil {
		return fmt.Errorf("not connected to WebSocket server")
	}

	if c.Verbose {
		// Convert to JSON for logging
		jsonBytes, err := json.MarshalIndent(message, "", "  ")
		if err != nil {
			log.Printf("DEBUG: Sending message (failed to format for debug): %+v", message)
		} else {
			log.Printf("DEBUG: Sending JSON message: \n%s", string(jsonBytes))
		}
	}

	return c.Conn.WriteJSON(message)
}

// handleMessages handles incoming WebSocket messages
func (c *WebSocketClient) handleMessages() {
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if c.Verbose {
				log.Printf("Error reading message: %v", err)
			}
			c.mu.Lock()
			c.Connected = false
			c.mu.Unlock()
			return
		}

		// Debug: Log raw message
		if c.Verbose {
			log.Printf("DEBUG: Raw message received: %s", string(message))
		}

		// Parse message to get type
		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			if c.Verbose {
				log.Printf("Error parsing message: %v", err)
				log.Printf("Failed message content: %s", string(message))
			}
			continue
		}

		// Get message type
		msgType, ok := data["type"].(string)
		if !ok {
			if c.Verbose {
				log.Printf("Message has no type field: %+v", data)
			}
			continue
		}

		if c.Verbose {
			log.Printf("Received message of type: %s with content: %+v", msgType, data)
		}

		// Call handler for message type
		c.mu.Lock()
		handler, ok := c.Handlers[msgType]
		c.mu.Unlock()
		if ok {
			if err := handler(message); err != nil {
				if c.Verbose {
					log.Printf("Error handling message of type %s: %v", msgType, err)
				}
			} else if c.Verbose {
				log.Printf("Successfully handled message of type: %s", msgType)
			}
		} else if c.Verbose {
			log.Printf("No handler registered for message type: %s", msgType)
		}
	}
}

// SendMessage sends a message to the WebSocket server
func (c *WebSocketClient) SendMessage(msg Message) error {
	if msg.Timestamp == "" {
		msg.Timestamp = time.Now().Format(time.RFC3339)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshaling message: %w", err)
	}

	if c.Verbose {
		// Pretty print for debugging
		var prettyMsg bytes.Buffer
		if err := json.Indent(&prettyMsg, data, "", "  "); err != nil {
			log.Printf("DEBUG: Sending message (failed to format for debug): %+v", msg)
		} else {
			log.Printf("DEBUG: Sending message: \n%s", prettyMsg.String())
		}
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	return nil
}

// SendScreenshot sends a screenshot through the WebSocket connection
func (c *WebSocketClient) SendScreenshot(screenshotData, format string, width, height int, description string) error {
	msg := Message{
		Type:           ScreenshotMessage,
		Message:        description,
		Timestamp:      time.Now().Format(time.RFC3339),
		ScreenshotData: screenshotData,
		ImageFormat:    format,
		Width:          width,
		Height:         height,
		Metadata: map[string]any{
			"platform": runtime.GOOS,
			"arch":     runtime.GOARCH,
		},
	}

	return c.SendMessage(msg)
}
