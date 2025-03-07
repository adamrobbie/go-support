package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/adamrobbie/go-support/pkg/client"
	"github.com/adamrobbie/go-support/pkg/permissions"
	"github.com/adamrobbie/go-support/pkg/screenshot"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	WebSocketURL       string
	TSWebSocketURL     string // TypeScript WebSocket server URL
	Verbose            bool
	SkipPermissions    bool
	UseTypeScriptWS    bool   // Whether to use the TypeScript WebSocket server
	ScreenshotDir      string // Directory to save screenshots
	Interactive        bool
	AutoScreenshot     bool // Whether to automatically take screenshots
	ScreenshotInterval int  // Interval in seconds between automatic screenshots
}

// App represents the application
type App struct {
	Config             Config
	PermManager        permissions.Manager
	WSClient           *client.WebSocketClient
	Done               chan struct{}
	stopAutoScreenshot chan struct{} // Channel to stop automatic screenshots
	Interrupt          chan os.Signal
}

// Message types
const (
	MessageTypeClientInfo     = "clientInfo"
	MessageTypeScreenshot     = "screenshot"
	MessageTypeTakeScreenshot = "takeScreenshot"
)

// ScreenshotMessage represents a screenshot message to be sent to the server
type ScreenshotMessage struct {
	Type      string `json:"type"`
	ImageURL  string `json:"imageUrl"` // Base64 encoded image data
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Timestamp string `json:"timestamp"`
}

// ClientInfoMessage represents client information to be sent to the server
type ClientInfoMessage struct {
	Type     string `json:"type"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
}

func main() {
	// Parse command line flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	interactive := flag.Bool("interactive", false, "Enable interactive mode")
	skipPermissions := flag.Bool("skip-permissions", false, "Skip permission checks")
	useTypeScriptWS := flag.Bool("use-ts-ws", false, "Use TypeScript WebSocket server")
	screenshotDir := flag.String("screenshot-dir", os.Getenv("SCREENSHOT_DIR"), "Directory to save screenshots")
	autoScreenshot := flag.Bool("auto-screenshot", false, "Automatically take screenshots")
	screenshotInterval := flag.Int("screenshot-interval", 10, "Interval in seconds between automatic screenshots")
	flag.Parse()

	// Create configuration
	var config Config
	config.Verbose = *verbose
	config.Interactive = *interactive
	config.SkipPermissions = *skipPermissions
	config.UseTypeScriptWS = *useTypeScriptWS
	config.ScreenshotDir = *screenshotDir
	config.AutoScreenshot = *autoScreenshot
	config.ScreenshotInterval = *screenshotInterval

	// Load additional configuration from environment
	if err := loadConfig(&config); err != nil {
		log.Fatal(err)
	}

	// Set up signal handling
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Create and run the application
	app := NewApp(config, interrupt)
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// loadConfig loads the application configuration
func loadConfig(config *Config) error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get WebSocket URLs from environment if not set
	if config.WebSocketURL == "" {
		config.WebSocketURL = os.Getenv("WEBSOCKET_URL")
		if config.WebSocketURL == "" {
			config.WebSocketURL = "ws://localhost:8080/ws"
		}
	}

	if config.TSWebSocketURL == "" {
		config.TSWebSocketURL = os.Getenv("TS_WEBSOCKET_URL")
		if config.TSWebSocketURL == "" {
			config.TSWebSocketURL = "ws://localhost:3000"
		}
	}

	// Create screenshot directory if it doesn't exist
	if config.ScreenshotDir == "" {
		config.ScreenshotDir = "screenshots"
	}

	if err := os.MkdirAll(config.ScreenshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshot directory: %w", err)
	}

	return nil
}

// NewApp creates a new application instance
func NewApp(config Config, interrupt chan os.Signal) *App {
	return &App{
		Config:             config,
		Done:               make(chan struct{}),
		stopAutoScreenshot: make(chan struct{}),
		Interrupt:          interrupt,
	}
}

// Run runs the application
func (a *App) Run() error {
	// Check permissions
	if !a.Config.SkipPermissions {
		if err := a.checkPermissions(); err != nil {
			return err
		}
	}

	// Connect to WebSocket server
	if err := a.connectWebSocket(); err != nil {
		return err
	}

	// Start automatic screenshot capture if enabled
	if a.Config.AutoScreenshot {
		go a.startAutoScreenshot()
	}

	// Take a test screenshot
	if err := a.takeTestScreenshot(); err != nil {
		log.Printf("Warning: Failed to take test screenshot: %v", err)
	}

	// Run the event loop
	return a.eventLoop()
}

// takeTestScreenshot takes a test screenshot and saves it to the configured directory
func (a *App) takeTestScreenshot() error {
	log.Println("Taking a test screenshot...")

	// Capture a screenshot with high quality
	ss, err := screenshot.Capture(screenshot.High)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Generate a filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(a.Config.ScreenshotDir, fmt.Sprintf("screenshot-%s.png", timestamp))

	// Save the screenshot to file
	if err := ss.SaveToFile(filename); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	log.Printf("Screenshot saved to: %s", filename)
	log.Printf("Screenshot dimensions: %dx%d", ss.Width, ss.Height)

	return nil
}

// captureAndSendScreenshot captures a screenshot and sends it to the server
func (a *App) captureAndSendScreenshot(quality screenshot.Quality, description string) error {
	// Capture screenshot
	log.Println("Capturing screenshot...")
	ss, err := screenshot.Capture(quality)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}
	log.Printf("Screenshot captured: %dx%d", ss.Width, ss.Height)

	// Resize the image if it's too large
	maxWidth, maxHeight := 1280, 720
	if ss.Width > maxWidth || ss.Height > maxHeight {
		log.Println("Resizing screenshot...")
		err = ss.Resize(maxWidth, maxHeight)
		if err != nil {
			return fmt.Errorf("failed to resize screenshot: %w", err)
		}
		log.Printf("Screenshot resized to: %dx%d", ss.Width, ss.Height)
	}

	// Compress the image
	err = ss.Compress(75) // 75% quality
	if err != nil {
		return fmt.Errorf("failed to compress screenshot: %w", err)
	}

	// Create a data URL
	dataURL := ss.ToBase64DataURL()

	// Create and send the message
	message := ScreenshotMessage{
		Type:      MessageTypeScreenshot,
		ImageURL:  dataURL,
		Width:     ss.Width,
		Height:    ss.Height,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	log.Println("Sending screenshot to server...")
	return a.WSClient.SendJSON(message)
}

// captureRegionAndSendScreenshot captures a screenshot of a specific region and sends it to the server
func (a *App) captureRegionAndSendScreenshot(region screenshot.Region, quality screenshot.Quality, description string) error {
	// Capture region screenshot
	log.Printf("Capturing region screenshot at (%d,%d) with size %dx%d...", region.X, region.Y, region.Width, region.Height)
	ss, err := screenshot.CaptureRegion(region, quality)
	if err != nil {
		return fmt.Errorf("failed to capture region screenshot: %w", err)
	}
	log.Printf("Region screenshot captured: %dx%d", ss.Width, ss.Height)

	// Compress the image
	err = ss.Compress(75) // 75% quality
	if err != nil {
		return fmt.Errorf("failed to compress screenshot: %w", err)
	}

	// Create a data URL
	dataURL := ss.ToBase64DataURL()

	// Create and send the message
	message := ScreenshotMessage{
		Type:      MessageTypeScreenshot,
		ImageURL:  dataURL,
		Width:     ss.Width,
		Height:    ss.Height,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	log.Println("Sending region screenshot to server...")
	return a.WSClient.SendJSON(message)
}

// checkPermissions checks if the application has the required permissions
func (a *App) checkPermissions() error {
	// Create a new permission manager
	a.PermManager = permissions.NewManager()

	// Check if screen sharing permissions are granted
	status, err := a.PermManager.CheckPermission(permissions.ScreenShare)
	if err != nil {
		return fmt.Errorf("failed to check screen sharing permission: %w", err)
	}

	if status != permissions.Granted {
		log.Println("Screen sharing permission not granted")

		// Request permission
		status, err := a.PermManager.RequestPermission(permissions.ScreenShare)
		if err != nil {
			return fmt.Errorf("failed to request screen sharing permission: %w", err)
		}

		// Check again after request
		if status != permissions.Granted {
			// For macOS, provide an interactive retry mechanism
			if runtime.GOOS == "darwin" {
				log.Println("Please grant screen recording permission in System Preferences")
				log.Println("Press 'r' to retry or 'q' to quit")

				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					input := scanner.Text()
					if input == "r" {
						status, _ := a.PermManager.CheckPermission(permissions.ScreenShare)
						if status == permissions.Granted {
							log.Println("Permission granted!")
							break
						} else {
							log.Println("Permission still not granted")
							log.Println("Press 'r' to retry or 'q' to quit")
						}
					} else if input == "q" {
						return fmt.Errorf("user quit due to permission not granted")
					}
				}
			} else {
				return fmt.Errorf("screen sharing permission not granted")
			}
		}
	}

	return nil
}

// connectWebSocket connects to the WebSocket server
func (a *App) connectWebSocket() error {
	var url string
	if a.Config.UseTypeScriptWS {
		url = a.Config.TSWebSocketURL
	} else {
		url = a.Config.WebSocketURL
	}

	log.Printf("Connecting to WebSocket server at %s...", url)
	a.WSClient = client.NewWebSocketClient(url, a.Config.Verbose)

	// Register message handlers
	a.WSClient.RegisterHandler(MessageTypeTakeScreenshot, func(data []byte) error {
		log.Println("Received screenshot request from server")
		return a.captureAndSendScreenshot(screenshot.High, "Requested screenshot")
	})

	// Connect to the server
	if err := a.WSClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	log.Println("Connected to WebSocket server")

	// Send client information after connection
	if err := a.sendClientInfo(); err != nil {
		log.Printf("Failed to send client info: %v", err)
	}

	return nil
}

// startAutoScreenshot starts a goroutine that takes screenshots at regular intervals
func (a *App) startAutoScreenshot() {
	ticker := time.NewTicker(time.Duration(a.Config.ScreenshotInterval) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting automatic screenshots every %d seconds", a.Config.ScreenshotInterval)

	for {
		select {
		case <-ticker.C:
			if a.WSClient != nil && a.WSClient.IsConnected() {
				log.Println("Taking automatic screenshot...")
				err := a.captureAndSendScreenshot(screenshot.High, "Automatic screenshot")
				if err != nil {
					log.Printf("Error taking automatic screenshot: %v", err)
				}
			}
		case <-a.stopAutoScreenshot:
			log.Println("Stopping automatic screenshots")
			return
		}
	}
}

// eventLoop handles the main event loop
func (a *App) eventLoop() error {
	// Set up interactive mode if enabled
	if a.Config.Interactive {
		scanner := bufio.NewScanner(os.Stdin)
		go a.handleUserInput(scanner)
	}

	// Clean up when done
	defer func() {
		if a.Config.AutoScreenshot {
			close(a.stopAutoScreenshot)
		}
		if a.WSClient != nil {
			a.WSClient.Close()
		}
	}()

	// Main event loop
	for {
		select {
		case <-a.Done:
			log.Println("Done signal received, shutting down...")
			return nil

		case <-a.Interrupt:
			log.Println("Interrupt received, shutting down...")
			return nil
		}
	}
}

// handleUserInput handles user input in interactive mode
func (a *App) handleUserInput(scanner *bufio.Scanner) {
	fmt.Println("\nInteractive mode enabled. Available commands:")
	fmt.Println("  screenshot - Capture and send a screenshot")
	fmt.Println("  region - Capture and send a region screenshot")
	fmt.Println("  exit - Exit the application")
	fmt.Println("Enter command:")

	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)

		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		switch command {
		case "screenshot":
			log.Println("Taking screenshot...")
			err := a.captureAndSendScreenshot(screenshot.High, "Manual screenshot")
			if err != nil {
				log.Printf("Error taking screenshot: %v", err)
			} else {
				log.Println("Screenshot sent successfully")
			}

		case "region":
			if len(parts) < 5 {
				log.Println("Usage: region <x> <y> <width> <height>")
				continue
			}

			// Parse region parameters
			x, y, width, height, err := parseRegionParams(parts[1:])
			if err != nil {
				log.Printf("Error parsing region parameters: %v", err)
				continue
			}

			log.Printf("Taking region screenshot at (%d,%d) with size %dx%d...", x, y, width, height)
			region := screenshot.Region{X: x, Y: y, Width: width, Height: height}
			err = a.captureRegionAndSendScreenshot(region, screenshot.High, "Manual region screenshot")
			if err != nil {
				log.Printf("Error taking region screenshot: %v", err)
			} else {
				log.Println("Region screenshot sent successfully")
			}

		case "exit":
			a.Interrupt <- os.Interrupt
			return

		default:
			fmt.Println("Unknown command. Available commands:")
			fmt.Println("  screenshot - Capture and send a screenshot")
			fmt.Println("  region - Capture and send a region screenshot")
			fmt.Println("  exit - Exit the application")
		}

		fmt.Println("Enter command:")
	}
}

// parseRegionParams parses region parameters from command line arguments
func parseRegionParams(args []string) (x, y, width, height int, err error) {
	if len(args) < 4 {
		return 0, 0, 0, 0, fmt.Errorf("not enough parameters")
	}

	x, err = strconv.Atoi(args[0])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid x coordinate: %w", err)
	}

	y, err = strconv.Atoi(args[1])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid y coordinate: %w", err)
	}

	width, err = strconv.Atoi(args[2])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid width: %w", err)
	}

	height, err = strconv.Atoi(args[3])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid height: %w", err)
	}

	return x, y, width, height, nil
}

// sendClientInfo sends information about the client to the server
func (a *App) sendClientInfo() error {
	message := ClientInfoMessage{
		Type:     MessageTypeClientInfo,
		Platform: runtime.GOOS,
		Version:  "1.0.0", // Your app version
	}

	return a.WSClient.SendJSON(message)
}

// handleWebSocketMessages processes incoming WebSocket messages
func handleWebSocketMessages(wsConn *websocket.Conn, done chan struct{}, config *Config) {
	defer close(done)

	for {
		var message map[string]interface{}
		err := wsConn.ReadJSON(&message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if msgType, ok := message["type"].(string); ok {
			if config.Verbose {
				log.Printf("Received message of type: %s", msgType)
			}

			switch msgType {
			case "welcome":
				log.Println("Received welcome message")

			case MessageTypeTakeScreenshot:
				log.Println("Received screenshot request from server")
				err := captureAndSendScreenshot(wsConn)
				if err != nil {
					log.Printf("Error sending screenshot: %v", err)
				} else {
					log.Println("Screenshot sent successfully")
				}

			// ... handle other message types ...

			default:
				if config.Verbose {
					log.Printf("Unhandled message type: %s", msgType)
				}
			}
		}
	}
}

// captureAndSendScreenshot captures a screenshot and sends it to the server
func captureAndSendScreenshot(wsConn *websocket.Conn) error {
	// Capture screenshot
	log.Println("Capturing screenshot...")
	img, err := screenshot.CaptureScreen()
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}
	log.Printf("Screenshot captured: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())

	// Resize the image if it's too large
	maxWidth, maxHeight := 1280, 720
	if img.Bounds().Dx() > maxWidth || img.Bounds().Dy() > maxHeight {
		log.Println("Resizing screenshot...")
		img = screenshot.ResizeImage(img, maxWidth, maxHeight)
		log.Printf("Screenshot resized to: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}

	// Encode as PNG
	log.Println("Encoding screenshot...")
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Encode as base64
	base64Data := base64.StdEncoding.EncodeToString(buf.Bytes())
	log.Printf("Screenshot encoded, size: %d bytes", len(base64Data))

	// Create a data URL
	dataURL := "data:image/png;base64," + base64Data

	// Create and send the message
	message := ScreenshotMessage{
		Type:      MessageTypeScreenshot,
		ImageURL:  dataURL,
		Width:     img.Bounds().Dx(),
		Height:    img.Bounds().Dy(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	log.Println("Sending screenshot to server...")
	return wsConn.WriteJSON(message)
}
