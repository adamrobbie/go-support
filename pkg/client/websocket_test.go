package client

import (
	"net"
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

	if client.Handlers == nil {
		t.Error("Expected Handlers map to be non-nil")
	}

	if client.ConnectTimeout != 10*time.Second {
		t.Errorf("Expected ConnectTimeout to be 10 seconds, got %v", client.ConnectTimeout)
	}

	if client.Connected {
		t.Error("Expected Connected to be false initially")
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

// TestRegisterHandler tests the RegisterHandler method
func TestRegisterHandler(t *testing.T) {
	client := NewWebSocketClient("ws://example.com", false)

	// Define a test handler
	testHandler := func(data []byte) error {
		return nil
	}

	// Register the handler
	client.RegisterHandler("test", testHandler)

	// Check that the handler was registered correctly
	client.mu.Lock()
	defer client.mu.Unlock()

	handler, exists := client.Handlers["test"]
	if !exists {
		t.Error("Handler was not registered")
	}

	// Check that the handler is the same function
	// We can't directly compare functions in Go, so we'll just check it's not nil
	if handler == nil {
		t.Error("Registered handler is nil")
	}
}

// TestIsConnected tests the IsConnected method
func TestIsConnected(t *testing.T) {
	client := NewWebSocketClient("ws://example.com", false)

	// Initially, the client should not be connected
	if client.IsConnected() {
		t.Error("New client should not be connected")
	}

	// Manually set the connected flag
	client.mu.Lock()
	client.Connected = true
	client.mu.Unlock()

	// Now the client should report as connected
	if !client.IsConnected() {
		t.Error("Client should report as connected")
	}
}

// mockConn is a mock implementation of the websocket.Conn interface for testing
type mockConn struct {
	writeCalled bool
	writeData   []byte
	writeType   int
	closeErr    error
	writeErr    error
}

func (m *mockConn) ReadMessage() (messageType int, p []byte, err error) {
	return 0, nil, nil
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	m.writeCalled = true
	m.writeData = data
	m.writeType = messageType
	return m.writeErr
}

func (m *mockConn) WriteJSON(v interface{}) error {
	return nil
}

func (m *mockConn) Close() error {
	return m.closeErr
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9090}
}

// TestSendJSON tests the SendJSON method
func TestSendJSON(t *testing.T) {
	client := NewWebSocketClient("ws://example.com", false)

	// Test error case: not connected
	message := map[string]interface{}{
		"type": "test",
		"data": "test data",
	}

	err := client.SendJSON(message)
	if err == nil {
		t.Error("SendJSON() should return an error when not connected")
	}

	// Check the error message
	expectedErrMsg := "not connected to WebSocket server"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	client := NewWebSocketClient("ws://example.com", false)

	// Test closing a client that's not connected
	err := client.Close()
	if err != nil {
		t.Errorf("Close() returned an error for a client that's not connected: %v", err)
	}

	// We can't easily test the successful case without a real WebSocket connection
	// or a more complex mock, but we've at least tested the error case
}
