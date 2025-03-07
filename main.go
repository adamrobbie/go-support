package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adamrobbie/go-support/client"
	"github.com/adamrobbie/go-support/pkg/appid"
	"github.com/adamrobbie/go-support/pkg/permissions"
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	WebSocketURL    string
	TSWebSocketURL  string // TypeScript WebSocket server URL
	Verbose         bool
	SkipPermissions bool
	UseTypeScriptWS bool // Whether to use the TypeScript WebSocket server
}

// App represents the application
type App struct {
	Config      Config
	PermManager permissions.Manager
	WSClient    *client.WebSocketClient
	Done        chan struct{}
	Interrupt   chan os.Signal
}

func main() {
	// Set up application identifier
	if err := appid.SetupAppIdentifier(); err != nil {
		log.Printf("Warning: Failed to set up application identifier: %v", err)
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create and run the application
	app := NewApp(config)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// loadConfig loads the application configuration from environment variables and command line flags
func loadConfig() (Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Get WebSocket URL from environment variable
	wsURL := os.Getenv("WEBSOCKET_URL")
	if wsURL == "" {
		return Config{}, fmt.Errorf("WEBSOCKET_URL environment variable not set")
	}

	// Get TypeScript WebSocket URL from environment variable (default to localhost:8080)
	tsWsURL := os.Getenv("TS_WEBSOCKET_URL")
	if tsWsURL == "" {
		tsWsURL = "ws://localhost:8080"
	}

	// Parse command line flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	skipPermissions := flag.Bool("skip-permissions", false, "Skip permission checks")
	useTypeScriptWS := flag.Bool("use-ts-ws", false, "Use TypeScript WebSocket server")
	flag.Parse()

	return Config{
		WebSocketURL:    wsURL,
		TSWebSocketURL:  tsWsURL,
		Verbose:         *verbose,
		SkipPermissions: *skipPermissions,
		UseTypeScriptWS: *useTypeScriptWS,
	}, nil
}

// NewApp creates a new application instance
func NewApp(config Config) *App {
	return &App{
		Config:      config,
		PermManager: permissions.NewManager(),
		Done:        make(chan struct{}),
		Interrupt:   make(chan os.Signal, 1),
	}
}

// Run runs the application
func (a *App) Run() error {
	// Log verbose mode if enabled
	if a.Config.Verbose {
		log.Println("Verbose mode enabled")
	}

	// Check permissions if not skipped
	if err := a.checkPermissions(); err != nil {
		return err
	}

	// Connect to WebSocket
	if err := a.connectWebSocket(); err != nil {
		return err
	}

	// Set up signal handling
	signal.Notify(a.Interrupt, os.Interrupt, syscall.SIGTERM)

	// Main event loop
	return a.eventLoop()
}

// checkPermissions checks if the required permissions are granted
func (a *App) checkPermissions() error {
	if a.Config.SkipPermissions {
		log.Println("Skipping permission checks as requested")
		return nil
	}

	log.Println("Checking screen sharing permission...")
	status, err := a.PermManager.RequestPermission(permissions.ScreenShare)
	if err != nil {
		log.Printf("Warning: Failed to request screen sharing permission: %v", err)
		log.Println("Continuing without screen sharing permission...")
		return nil
	}

	switch status {
	case permissions.Granted:
		log.Println("Screen sharing permission granted")
	case permissions.Denied:
		log.Println("Screen sharing permission denied")
		log.Println("The application may not function correctly without screen sharing permission")
		log.Println("Continuing anyway...")
	case permissions.Requested:
		// If permission was requested but the user chose to quit, exit gracefully
		log.Println("Exiting application. Please restart after granting permission.")
		return fmt.Errorf("permission requested, restart required")
	default:
		log.Println("Screen sharing permission status unknown")
		log.Println("Continuing anyway...")
	}

	return nil
}

// connectWebSocket connects to the WebSocket server
func (a *App) connectWebSocket() error {
	var wsURL string

	if a.Config.UseTypeScriptWS {
		wsURL = a.Config.TSWebSocketURL
		log.Printf("Using TypeScript WebSocket server at %s", wsURL)

		// Create and connect the WebSocket client
		a.WSClient = client.NewWebSocketClient(wsURL, a.Config.Verbose)
		return a.WSClient.Connect()
	} else {
		wsURL = a.Config.WebSocketURL
		log.Printf("Using standard WebSocket server at %s", wsURL)

		// Use the original WebSocket connection logic
		log.Printf("Connecting to WebSocket at %s", wsURL)

		// Create and connect the WebSocket client (but use it as a regular client)
		a.WSClient = client.NewWebSocketClient(wsURL, a.Config.Verbose)
		return a.WSClient.Connect()
	}
}

// eventLoop handles the main event loop
func (a *App) eventLoop() error {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	// Clean up when done
	defer func() {
		if a.WSClient != nil {
			a.WSClient.Close()
		}
	}()

	for {
		select {
		case msg := <-a.WSClient.Receive:
			// Handle received messages
			fmt.Printf("Received message: %+v\n", msg)

		case <-ticker.C:
			// Send a chat message periodically
			if a.Config.UseTypeScriptWS {
				a.WSClient.Send <- client.Message{
					Type:    client.ChatMessage,
					Message: "Hello from Go client!",
				}
			}

		case <-a.Interrupt:
			log.Println("Interrupt received, shutting down...")
			return nil
		}
	}
}
