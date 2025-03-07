package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("WEBSOCKET_URL", "wss://test.example.com/ws")
	os.Setenv("TS_WEBSOCKET_URL", "ws://localhost:9090")

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() returned an error: %v", err)
	}

	// Check that the configuration was loaded correctly
	if config.WebSocketURL != "wss://test.example.com/ws" {
		t.Errorf("Expected WebSocketURL to be 'wss://test.example.com/ws', got '%s'", config.WebSocketURL)
	}

	if config.TSWebSocketURL != "ws://localhost:9090" {
		t.Errorf("Expected TSWebSocketURL to be 'ws://localhost:9090', got '%s'", config.TSWebSocketURL)
	}

	// Default values should be false
	if config.Verbose {
		t.Error("Expected Verbose to be false by default")
	}

	if config.SkipPermissions {
		t.Error("Expected SkipPermissions to be false by default")
	}

	if config.UseTypeScriptWS {
		t.Error("Expected UseTypeScriptWS to be false by default")
	}
}

func TestNewApp(t *testing.T) {
	// Create a test configuration
	config := Config{
		WebSocketURL:    "wss://test.example.com/ws",
		TSWebSocketURL:  "ws://localhost:9090",
		Verbose:         true,
		SkipPermissions: true,
		UseTypeScriptWS: true,
	}

	// Create a new app
	app := NewApp(config)

	// Check that the app was created correctly
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	if app.Config.WebSocketURL != config.WebSocketURL {
		t.Errorf("Expected WebSocketURL to be '%s', got '%s'", config.WebSocketURL, app.Config.WebSocketURL)
	}

	if app.Config.TSWebSocketURL != config.TSWebSocketURL {
		t.Errorf("Expected TSWebSocketURL to be '%s', got '%s'", config.TSWebSocketURL, app.Config.TSWebSocketURL)
	}

	if app.Config.Verbose != config.Verbose {
		t.Errorf("Expected Verbose to be %v, got %v", config.Verbose, app.Config.Verbose)
	}

	if app.Config.SkipPermissions != config.SkipPermissions {
		t.Errorf("Expected SkipPermissions to be %v, got %v", config.SkipPermissions, app.Config.SkipPermissions)
	}

	if app.Config.UseTypeScriptWS != config.UseTypeScriptWS {
		t.Errorf("Expected UseTypeScriptWS to be %v, got %v", config.UseTypeScriptWS, app.Config.UseTypeScriptWS)
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
