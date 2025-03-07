package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adamrobbie/go-support/pkg/appid"
	"github.com/adamrobbie/go-support/pkg/permissions"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	WebSocketURL    string
	Verbose         bool
	SkipPermissions bool
}

// App represents the application
type App struct {
	Config      Config
	PermManager permissions.Manager
	WSConn      *websocket.Conn
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

	// Parse command line flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	skipPermissions := flag.Bool("skip-permissions", false, "Skip permission checks")
	flag.Parse()

	return Config{
		WebSocketURL:    wsURL,
		Verbose:         *verbose,
		SkipPermissions: *skipPermissions,
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
	defer a.WSConn.Close()

	// Set up signal handling
	signal.Notify(a.Interrupt, os.Interrupt, syscall.SIGTERM)

	// Start reading messages
	go a.readMessages()

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
	log.Printf("Connecting to WebSocket at %s", a.Config.WebSocketURL)

	conn, _, err := websocket.DefaultDialer.Dial(a.Config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %w", err)
	}

	a.WSConn = conn
	log.Println("Connected to WebSocket server")
	return nil
}

// readMessages reads messages from the WebSocket connection
func (a *App) readMessages() {
	defer close(a.Done)
	for {
		_, message, err := a.WSConn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			return
		}
		fmt.Printf("Received: %s\n", message)
	}
}

// eventLoop handles the main event loop
func (a *App) eventLoop() error {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-a.Done:
			return nil
		case <-ticker.C:
			if err := a.sendPing(); err != nil {
				return err
			}
		case <-a.Interrupt:
			return a.gracefulShutdown()
		}
	}
}

// sendPing sends a ping message to the WebSocket server
func (a *App) sendPing() error {
	err := a.WSConn.WriteMessage(websocket.TextMessage, []byte("ping"))
	if err != nil {
		return fmt.Errorf("error sending ping: %w", err)
	}
	if a.Config.Verbose {
		log.Println("Sent ping")
	}
	return nil
}

// gracefulShutdown performs a graceful shutdown of the application
func (a *App) gracefulShutdown() error {
	log.Println("Interrupt received, closing connection...")

	// Cleanly close the connection by sending a close message
	err := a.WSConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("error during closing WebSocket: %w", err)
	}

	// Wait for the server to close the connection
	select {
	case <-a.Done:
	case <-time.After(time.Second):
	}
	return nil
}
