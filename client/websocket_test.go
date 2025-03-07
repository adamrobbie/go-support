package client

import (
	"testing"
	"time"
)

func TestNewWebSocketClient(t *testing.T) {
	// Create a new WebSocket client
	client := NewWebSocketClient("ws://example.com", true)

	// Check that the client was created correctly
	if client == nil {
		t.Fatal("NewWebSocketClient() returned nil")
	}

	if client.URL != "ws://example.com" {
		t.Errorf("Expected URL to be 'ws://example.com', got '%s'", client.URL)
	}

	if !client.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if client.Send == nil {
		t.Error("Expected Send channel to be non-nil")
	}

	if client.Receive == nil {
		t.Error("Expected Receive channel to be non-nil")
	}

	if client.Done == nil {
		t.Error("Expected Done channel to be non-nil")
	}

	if client.Interrupt == nil {
		t.Error("Expected Interrupt channel to be non-nil")
	}

	if client.ReconnectWait != 5*time.Second {
		t.Errorf("Expected ReconnectWait to be 5 seconds, got %v", client.ReconnectWait)
	}
}

func TestMessageTypes(t *testing.T) {
	// Test message types
	if ChatMessage != "chat" {
		t.Errorf("Expected ChatMessage to be 'chat', got '%s'", ChatMessage)
	}

	if PingMessage != "ping" {
		t.Errorf("Expected PingMessage to be 'ping', got '%s'", PingMessage)
	}

	if CustomMessage != "custom" {
		t.Errorf("Expected CustomMessage to be 'custom', got '%s'", CustomMessage)
	}
}

func TestMessage(t *testing.T) {
	// Create a message
	msg := Message{
		Type:      ChatMessage,
		Message:   "Hello, world!",
		Timestamp: "2023-01-01T00:00:00Z",
		Metadata: map[string]any{
			"key": "value",
		},
		Platform: "test",
		Version:  "1.0.0",
		Extra: map[string]any{
			"extra": "data",
		},
	}

	// Check that the message was created correctly
	if msg.Type != ChatMessage {
		t.Errorf("Expected Type to be '%s', got '%s'", ChatMessage, msg.Type)
	}

	if msg.Message != "Hello, world!" {
		t.Errorf("Expected Message to be 'Hello, world!', got '%s'", msg.Message)
	}

	if msg.Timestamp != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Timestamp to be '2023-01-01T00:00:00Z', got '%s'", msg.Timestamp)
	}

	if msg.Metadata["key"] != "value" {
		t.Errorf("Expected Metadata['key'] to be 'value', got '%v'", msg.Metadata["key"])
	}

	if msg.Platform != "test" {
		t.Errorf("Expected Platform to be 'test', got '%s'", msg.Platform)
	}

	if msg.Version != "1.0.0" {
		t.Errorf("Expected Version to be '1.0.0', got '%s'", msg.Version)
	}

	if msg.Extra["extra"] != "data" {
		t.Errorf("Expected Extra['extra'] to be 'data', got '%v'", msg.Extra["extra"])
	}
}
