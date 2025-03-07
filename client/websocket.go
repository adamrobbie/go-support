package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/adamrobbie/go-support/pkg/appid"
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
)

// Message represents a message to be sent to the WebSocket server
type Message struct {
	Type      MessageType    `json:"type"`
	Message   string         `json:"message,omitempty"`
	Timestamp string         `json:"timestamp,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Platform  string         `json:"platform,omitempty"`
	Version   string         `json:"version,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

// WebSocketClient represents a WebSocket client
type WebSocketClient struct {
	URL           string
	Conn          *websocket.Conn
	Send          chan Message
	Receive       chan Message
	Done          chan struct{}
	Interrupt     chan os.Signal
	ReconnectWait time.Duration
	Verbose       bool
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(url string, verbose bool) *WebSocketClient {
	return &WebSocketClient{
		URL:           url,
		Send:          make(chan Message),
		Receive:       make(chan Message),
		Done:          make(chan struct{}),
		Interrupt:     make(chan os.Signal, 1),
		ReconnectWait: 5 * time.Second,
		Verbose:       verbose,
	}
}

// Connect connects to the WebSocket server
func (c *WebSocketClient) Connect() error {
	if c.Verbose {
		log.Printf("Connecting to WebSocket server at %s", c.URL)
	}

	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(c.URL, nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %w", err)
	}

	c.Conn = conn

	// Set up signal handling
	signal.Notify(c.Interrupt, os.Interrupt)

	// Send initial message with client information
	initialMsg := Message{
		Type:    CustomMessage,
		Message: "Go client connected",
		Metadata: map[string]any{
			"platform": runtime.GOOS,
			"arch":     runtime.GOARCH,
			"version":  appid.AppVersion,
			"appId":    appid.AppID,
			"appName":  appid.AppName,
		},
	}

	err = c.SendMessage(initialMsg)
	if err != nil {
		return fmt.Errorf("error sending initial message: %w", err)
	}

	// Start goroutines for reading and writing
	go c.readPump()
	go c.writePump()

	return nil
}

// readPump pumps messages from the WebSocket connection to the Receive channel
func (c *WebSocketClient) readPump() {
	defer func() {
		c.Conn.Close()
		close(c.Done)
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if c.Verbose {
			log.Printf("Received message: %s", string(message))
		}

		// Handle ping messages automatically
		if msg.Type == PingMessage {
			pongMsg := Message{
				Type:      "pong",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			c.Send <- pongMsg
		} else {
			c.Receive <- msg
		}
	}
}

// writePump pumps messages from the Send channel to the WebSocket connection
func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			err = c.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}

			if c.Verbose {
				log.Printf("Sent message: %s", string(data))
			}

		case <-ticker.C:
			// Send ping message
			pingMsg := Message{
				Type:      PingMessage,
				Timestamp: time.Now().Format(time.RFC3339),
			}
			c.Send <- pingMsg

		case <-c.Interrupt:
			// Cleanly close the connection
			err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Error during closing WebSocket: %v", err)
				return
			}
			select {
			case <-c.Done:
			case <-time.After(time.Second):
			}
			return
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

	err = c.Conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	if c.Verbose {
		log.Printf("Sent message: %s", string(data))
	}

	return nil
}

// Close closes the WebSocket connection
func (c *WebSocketClient) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}
