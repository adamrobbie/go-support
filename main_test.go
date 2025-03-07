package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/adamrobbie/go-support/pkg/permissions"
	"github.com/gorilla/websocket"
)

// MockPermissionManager is a mock implementation of the permissions.Manager interface
type MockPermissionManager struct {
	RequestPermissionFunc func(permType permissions.PermissionType) (permissions.PermissionStatus, error)
	CheckPermissionFunc   func(permType permissions.PermissionType) (permissions.PermissionStatus, error)
}

func (m *MockPermissionManager) RequestPermission(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
	if m.RequestPermissionFunc != nil {
		return m.RequestPermissionFunc(permType)
	}
	return permissions.Unknown, nil
}

func (m *MockPermissionManager) CheckPermission(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
	if m.CheckPermissionFunc != nil {
		return m.CheckPermissionFunc(permType)
	}
	return permissions.Unknown, nil
}

// TestConfig tests the Config struct
func TestConfig(t *testing.T) {
	config := Config{
		WebSocketURL:    "ws://example.com",
		Verbose:         true,
		SkipPermissions: true,
	}

	if config.WebSocketURL != "ws://example.com" {
		t.Errorf("Expected WebSocketURL to be 'ws://example.com', got '%s'", config.WebSocketURL)
	}

	if !config.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if !config.SkipPermissions {
		t.Error("Expected SkipPermissions to be true")
	}
}

// TestNewApp tests the NewApp function
func TestNewApp(t *testing.T) {
	config := Config{
		WebSocketURL:    "ws://example.com",
		Verbose:         true,
		SkipPermissions: true,
	}

	app := NewApp(config)

	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.Config.WebSocketURL != config.WebSocketURL {
		t.Errorf("Expected WebSocketURL to be '%s', got '%s'", config.WebSocketURL, app.Config.WebSocketURL)
	}

	if app.Config.Verbose != config.Verbose {
		t.Errorf("Expected Verbose to be %v, got %v", config.Verbose, app.Config.Verbose)
	}

	if app.Config.SkipPermissions != config.SkipPermissions {
		t.Errorf("Expected SkipPermissions to be %v, got %v", config.SkipPermissions, app.Config.SkipPermissions)
	}

	if app.PermManager == nil {
		t.Error("Expected PermManager to be non-nil")
	}

	if app.Done == nil {
		t.Error("Expected Done channel to be non-nil")
	}

	if app.Interrupt == nil {
		t.Error("Expected Interrupt channel to be non-nil")
	}
}

// TestCheckPermissions tests the checkPermissions method
func TestCheckPermissions(t *testing.T) {
	// Test with skip permissions
	app := &App{
		Config: Config{
			SkipPermissions: true,
		},
	}

	err := app.checkPermissions()
	if err != nil {
		t.Errorf("checkPermissions() with SkipPermissions=true returned an error: %v", err)
	}

	// Test with permission granted
	mockManager := &MockPermissionManager{
		RequestPermissionFunc: func(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
			return permissions.Granted, nil
		},
	}

	app = &App{
		Config:      Config{SkipPermissions: false},
		PermManager: mockManager,
	}

	err = app.checkPermissions()
	if err != nil {
		t.Errorf("checkPermissions() with Granted permission returned an error: %v", err)
	}

	// Test with permission denied
	mockManager = &MockPermissionManager{
		RequestPermissionFunc: func(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
			return permissions.Denied, nil
		},
	}

	app = &App{
		Config:      Config{SkipPermissions: false},
		PermManager: mockManager,
	}

	err = app.checkPermissions()
	if err != nil {
		t.Errorf("checkPermissions() with Denied permission returned an error: %v", err)
	}

	// Test with permission requested
	mockManager = &MockPermissionManager{
		RequestPermissionFunc: func(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
			return permissions.Requested, nil
		},
	}

	app = &App{
		Config:      Config{SkipPermissions: false},
		PermManager: mockManager,
	}

	err = app.checkPermissions()
	if err == nil {
		t.Error("checkPermissions() with Requested permission should return an error")
	}

	// Test with permission error
	mockManager = &MockPermissionManager{
		RequestPermissionFunc: func(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
			return permissions.Unknown, errors.New("permission error")
		},
	}

	app = &App{
		Config:      Config{SkipPermissions: false},
		PermManager: mockManager,
	}

	err = app.checkPermissions()
	if err != nil {
		t.Errorf("checkPermissions() with permission error should not return an error, got: %v", err)
	}
}

// TestWebSocketConnection tests the WebSocket connection functionality
func TestWebSocketConnection(t *testing.T) {
	// Create a test WebSocket server
	var upgrader = websocket.Upgrader{}

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		// Echo all messages back to the client
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Convert http://127.0.0.1... to ws://127.0.0.1...
	wsURL := "ws" + server.URL[4:]

	// Create an app with the test server URL
	app := &App{
		Config: Config{
			WebSocketURL:    wsURL,
			Verbose:         true,
			SkipPermissions: true,
		},
		Done:      make(chan struct{}),
		Interrupt: make(chan os.Signal, 1),
	}

	// Connect to the WebSocket server
	err := app.connectWebSocket()
	if err != nil {
		t.Fatalf("connectWebSocket() returned an error: %v", err)
	}
	defer app.WSConn.Close()

	// Test sending a ping
	err = app.sendPing()
	if err != nil {
		t.Errorf("sendPing() returned an error: %v", err)
	}

	// Start reading messages
	go app.readMessages()

	// Wait a moment for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Test graceful shutdown
	err = app.gracefulShutdown()
	if err != nil {
		t.Errorf("gracefulShutdown() returned an error: %v", err)
	}
}
