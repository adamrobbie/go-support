package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
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
	"github.com/adamrobbie/go-support/pkg/remote"
	"github.com/adamrobbie/go-support/pkg/screenshot"
	"github.com/adamrobbie/go-support/pkg/video"
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
	TestRobotgo        bool // Whether to run RobotGo tests
	RequestPermissions bool // Whether to explicitly request permissions

	// Video streaming options
	VideoStreaming    bool   // Whether to enable video streaming
	VideoQuality      string // Quality of the video stream (low, medium, high)
	VideoFPS          int    // Frames per second for video streaming
	VideoRecording    bool   // Whether to enable video recording
	VideoRecordingDir string // Directory to save video recordings
}

// App represents the application
type App struct {
	Config             Config
	PermManager        permissions.Manager
	WSClient           *client.WebSocketClient
	Done               chan struct{}
	stopAutoScreenshot chan struct{} // Channel to stop automatic screenshots
	Interrupt          chan os.Signal
	RemoteController   *remote.RemoteController
	VideoStream        *video.VideoStream
}

// Message types
const (
	MessageTypeClientInfo            = "clientInfo"
	MessageTypeScreenshot            = "screenshot"
	MessageTypeTakeScreenshot        = "takeScreenshot"
	MessageTypeMouseEvent            = "mouseEvent"
	MessageTypeKeyboardEvent         = "keyboardEvent"
	MessageTypeScreenSize            = "screenSize"
	MessageTypeMousePosition         = "mousePosition"
	MessageTypeVideoFrame            = "videoFrame"
	MessageTypeStartVideo            = "startVideo"
	MessageTypeStopVideo             = "stopVideo"
	MessageTypeStartRecording        = "startRecording"
	MessageTypeStopRecording         = "stopRecording"
	MessageTypeScreenRecordingStatus = "screenRecordingStatus" // New message type for screen recording status
	MessageTypeScreenRecordingSaved  = "screenRecordingSaved"  // New message type for when recording is saved
	MessageTypeGetRecordingStatus    = "getRecordingStatus"    // New message type for requesting recording status
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

// dumpMessageTypes logs all available message types for debugging
func dumpMessageTypes() {
	log.Println("=== Available WebSocket Message Types ===")
	log.Printf("ClientInfo:          %s", MessageTypeClientInfo)
	log.Printf("Screenshot:          %s", MessageTypeScreenshot)
	log.Printf("TakeScreenshot:      %s", MessageTypeTakeScreenshot)
	log.Printf("MouseEvent:          %s", MessageTypeMouseEvent)
	log.Printf("KeyboardEvent:       %s", MessageTypeKeyboardEvent)
	log.Printf("ScreenSize:          %s", MessageTypeScreenSize)
	log.Printf("MousePosition:       %s", MessageTypeMousePosition)
	log.Printf("VideoFrame:          %s", MessageTypeVideoFrame)
	log.Printf("StartVideo:          %s", MessageTypeStartVideo)
	log.Printf("StopVideo:           %s", MessageTypeStopVideo)
	log.Printf("StartRecording:      %s", MessageTypeStartRecording)
	log.Printf("StopRecording:       %s", MessageTypeStopRecording)
	log.Printf("ScreenRecordingStatus: %s", MessageTypeScreenRecordingStatus)
	log.Printf("ScreenRecordingSaved:  %s", MessageTypeScreenRecordingSaved)
	log.Printf("GetRecordingStatus:    %s", MessageTypeGetRecordingStatus)
	log.Println("========================================")
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
	testRobotgo := flag.Bool("test-robotgo", false, "Test RobotGo functionality")
	requestPermissions := flag.Bool("request-permissions", false, "Explicitly request permissions")

	// Video streaming flags
	videoStreaming := flag.Bool("video-streaming", false, "Enable video streaming")
	videoQuality := flag.String("video-quality", "medium", "Quality of the video stream (low, medium, high)")
	videoFPS := flag.Int("video-fps", 10, "Frames per second for video streaming")
	videoRecording := flag.Bool("video-recording", false, "Enable video recording")
	videoRecordingDir := flag.String("video-recording-dir", "recordings", "Directory to save video recordings")

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
	config.TestRobotgo = *testRobotgo
	config.RequestPermissions = *requestPermissions

	// Video streaming configuration
	config.VideoStreaming = *videoStreaming
	config.VideoQuality = *videoQuality
	config.VideoFPS = *videoFPS
	config.VideoRecording = *videoRecording
	config.VideoRecordingDir = *videoRecordingDir

	// Load additional configuration from environment
	if err := loadConfig(&config); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Print configuration if verbose
	if config.Verbose {
		log.Printf("Configuration: %+v", config)
		// Dump message types for debugging
		dumpMessageTypes()
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
	log.Println("Starting go-support...")

	// Check permissions if not skipped
	if !a.Config.SkipPermissions {
		if err := a.checkPermissions(); err != nil {
			return fmt.Errorf("permission check failed: %w", err)
		}
	}

	// Explicitly request permissions if requested
	if a.Config.RequestPermissions {
		if err := a.requestPermissionsInteractive(); err != nil {
			return fmt.Errorf("permission request failed: %w", err)
		}
		return nil // Exit after requesting permissions
	}

	// Test RobotGo if requested
	if a.Config.TestRobotgo {
		if err := a.testRobotgo(); err != nil {
			return fmt.Errorf("RobotGo test failed: %w", err)
		}
		return nil // Exit after test
	}

	// Connect to WebSocket server
	if err := a.connectWebSocket(); err != nil {
		return fmt.Errorf("failed to connect to WebSocket server: %w", err)
	}

	// Initialize video streaming if enabled
	if a.Config.VideoStreaming {
		if err := a.initVideoStream(); err != nil {
			return fmt.Errorf("failed to initialize video stream: %w", err)
		}
	}

	// Start automatic screenshots if enabled
	if a.Config.AutoScreenshot {
		go a.startAutoScreenshot()
	}

	// Start event loop
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
	a.PermManager = permissions.NewManager(a.Config.Verbose)

	// Check if screen sharing permissions are granted
	log.Println("Checking screen sharing permissions...")
	screenShareStatus, err := a.PermManager.CheckPermission(permissions.ScreenShare)
	if err != nil {
		return fmt.Errorf("failed to check screen sharing permission: %w", err)
	}

	if screenShareStatus != permissions.Granted {
		log.Println("⚠️ Screen sharing permission not granted")

		// Ask user if they want to request permission interactively
		fmt.Println("\nScreen sharing permission is required for screenshot functionality.")
		fmt.Println("Would you like to grant this permission now? (y/n)")
		var input string
		fmt.Scanln(&input)

		if input == "y" || input == "Y" {
			// Request permission interactively
			granted := a.PermManager.RequestPermissionInteractive(permissions.ScreenShare)
			if !granted {
				return fmt.Errorf("screen sharing permission not granted")
			}
		} else {
			return fmt.Errorf("screen sharing permission not granted")
		}
	}

	log.Println("✅ Screen sharing permission granted")

	// Always check accessibility permissions for remote control
	log.Println("Checking accessibility permissions for remote control...")
	accessibilityStatus, err := a.PermManager.CheckPermission(permissions.RemoteControl)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to check accessibility permission: %v", err)
		log.Println("Remote control features may not work correctly")
		// Don't return error here, just warn the user
	} else if accessibilityStatus != permissions.Granted {
		log.Println("⚠️ Warning: Accessibility permission not granted")

		// Ask user if they want to request permission interactively
		fmt.Println("\nAccessibility permission is required for remote control functionality.")
		fmt.Println("Would you like to grant this permission now? (y/n)")
		var input string
		fmt.Scanln(&input)

		if input == "y" || input == "Y" {
			// Request permission interactively
			granted := a.PermManager.RequestPermissionInteractive(permissions.RemoteControl)
			if granted {
				log.Println("✅ Accessibility permission granted")
			} else {
				log.Println("⚠️ Warning: Accessibility permission not granted")
				log.Println("Remote control features may not work correctly")
			}
		} else {
			log.Println("⚠️ Warning: Accessibility permission not granted")
			log.Println("Remote control features may not work correctly")
		}
	} else {
		log.Println("✅ Accessibility permission granted")
	}

	return nil
}

// connectWebSocket connects to the WebSocket server
func (a *App) connectWebSocket() error {
	// Determine the WebSocket URL
	var url string
	if a.Config.UseTypeScriptWS {
		url = a.Config.TSWebSocketURL
	} else {
		url = a.Config.WebSocketURL
	}

	log.Printf("Connecting to WebSocket server at %s...", url)

	// Create a new WebSocket client
	a.WSClient = client.NewWebSocketClient(url, a.Config.Verbose)

	// Create a new remote controller
	a.RemoteController = remote.NewRemoteController(a.PermManager, a.Config.Verbose)

	// Register message handlers
	a.WSClient.RegisterHandler(MessageTypeTakeScreenshot, func(data []byte) error {
		log.Println("DEBUG: Received screenshot request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Screenshot request details: %+v", msg)
		}

		return a.captureAndSendScreenshot(screenshot.High, "Requested screenshot")
	})

	a.WSClient.RegisterHandler(MessageTypeMouseEvent, func(data []byte) error {
		log.Println("DEBUG: Received mouse event from server")

		var event remote.MouseEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("ERROR: Failed to parse mouse event: %v", err)
			log.Printf("ERROR: Raw mouse event data: %s", string(data))
			return fmt.Errorf("failed to parse mouse event: %w", err)
		}

		log.Printf("DEBUG: Mouse event details: %+v", event)
		return a.RemoteController.ExecuteMouseEvent(event)
	})

	a.WSClient.RegisterHandler(MessageTypeKeyboardEvent, func(data []byte) error {
		log.Println("DEBUG: Received keyboard event from server")

		var event remote.KeyboardEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("ERROR: Failed to parse keyboard event: %v", err)
			log.Printf("ERROR: Raw keyboard event data: %s", string(data))
			return fmt.Errorf("failed to parse keyboard event: %w", err)
		}

		log.Printf("DEBUG: Keyboard event details: %+v", event)
		return a.RemoteController.ExecuteKeyboardEvent(event)
	})

	a.WSClient.RegisterHandler(MessageTypeScreenSize, func(data []byte) error {
		log.Println("DEBUG: Received screen size request from server")

		width, height, err := a.RemoteController.GetScreenSize()
		if err != nil {
			log.Printf("ERROR: Failed to get screen size: %v", err)
			return fmt.Errorf("failed to get screen size: %w", err)
		}

		log.Printf("DEBUG: Sending screen size response: width=%d, height=%d", width, height)
		message := map[string]interface{}{
			"type":   MessageTypeScreenSize,
			"width":  width,
			"height": height,
		}

		return a.WSClient.SendJSON(message)
	})

	a.WSClient.RegisterHandler(MessageTypeMousePosition, func(data []byte) error {
		log.Println("DEBUG: Received mouse position request from server")

		x, y, err := a.RemoteController.GetMousePosition()
		if err != nil {
			log.Printf("ERROR: Failed to get mouse position: %v", err)
			return fmt.Errorf("failed to get mouse position: %w", err)
		}

		log.Printf("DEBUG: Sending mouse position response: x=%d, y=%d", x, y)
		message := map[string]interface{}{
			"type": MessageTypeMousePosition,
			"x":    x,
			"y":    y,
		}

		return a.WSClient.SendJSON(message)
	})

	// Register video streaming handlers
	a.WSClient.RegisterHandler(MessageTypeStartVideo, func(data []byte) error {
		log.Println("DEBUG: Received start video streaming request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Start video request details: %+v", msg)
		}

		err := a.startVideoStreaming()
		if err != nil {
			log.Printf("ERROR: Failed to start video streaming: %v", err)
		} else {
			log.Println("DEBUG: Video streaming started successfully")
		}
		return err
	})

	a.WSClient.RegisterHandler(MessageTypeStopVideo, func(data []byte) error {
		log.Println("DEBUG: Received stop video streaming request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Stop video request details: %+v", msg)
		}

		a.stopVideoStreaming()
		log.Println("DEBUG: Video streaming stopped successfully")
		return nil
	})

	a.WSClient.RegisterHandler(MessageTypeStartRecording, func(data []byte) error {
		log.Println("DEBUG: Received start video recording request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Start recording request details: %+v", msg)
		}

		err := a.startVideoRecording()
		if err != nil {
			log.Printf("ERROR: Failed to start video recording: %v", err)
		} else {
			log.Println("DEBUG: Video recording started successfully")
		}
		return err
	})

	a.WSClient.RegisterHandler(MessageTypeStopRecording, func(data []byte) error {
		log.Println("DEBUG: Received stop video recording request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Stop recording request details: %+v", msg)
		}

		err := a.stopVideoRecording()
		if err != nil {
			log.Printf("ERROR: Failed to stop video recording: %v", err)
		} else {
			log.Println("DEBUG: Video recording stopped successfully")
		}
		return err
	})

	// Register recording status request handler
	a.WSClient.RegisterHandler(MessageTypeGetRecordingStatus, func(data []byte) error {
		log.Println("DEBUG: Received recording status request from server")

		// Parse the full message for debugging
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err == nil && a.Config.Verbose {
			log.Printf("DEBUG: Recording status request details: %+v", msg)
		}

		err := a.getRecordingStatus()
		if err != nil {
			log.Printf("ERROR: Failed to get recording status: %v", err)
		} else {
			log.Println("DEBUG: Recording status sent successfully")
		}
		return err
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

	// Send screen size information
	width, height, err := a.RemoteController.GetScreenSize()
	if err != nil {
		log.Printf("Failed to get screen size: %v", err)
	} else {
		screenSizeMsg := map[string]interface{}{
			"type":   MessageTypeScreenSize,
			"width":  width,
			"height": height,
		}

		if err := a.WSClient.SendJSON(screenSizeMsg); err != nil {
			log.Printf("Failed to send screen size info: %v", err)
		}
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

// handleUserInput handles user input from the console
func (a *App) handleUserInput(scanner *bufio.Scanner) {
	for scanner.Scan() {
		input := scanner.Text()
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		command := args[0]
		switch command {
		case "exit", "quit":
			log.Println("Exiting...")
			close(a.Done)
			return
		case "screenshot":
			quality := screenshot.Medium
			if len(args) > 1 {
				switch args[1] {
				case "low":
					quality = screenshot.Low
				case "high":
					quality = screenshot.High
				}
			}
			if err := a.captureAndSendScreenshot(quality, "User-initiated screenshot"); err != nil {
				log.Printf("Error capturing screenshot: %v", err)
			}
		case "region":
			if len(args) < 5 {
				log.Println("Usage: region <x> <y> <width> <height>")
				continue
			}
			x, y, width, height, err := parseRegionParams(args[1:])
			if err != nil {
				log.Printf("Error parsing region parameters: %v", err)
				continue
			}
			region := screenshot.Region{X: x, Y: y, Width: width, Height: height}
			if err := a.captureRegionAndSendScreenshot(region, screenshot.High, "User-initiated region screenshot"); err != nil {
				log.Printf("Error capturing region screenshot: %v", err)
			}
		case "auto":
			if len(args) > 1 && args[1] == "off" {
				log.Println("Stopping automatic screenshots")
				a.stopAutoScreenshot <- struct{}{}
			} else {
				log.Printf("Starting automatic screenshots every %d seconds", a.Config.ScreenshotInterval)
				go a.startAutoScreenshot()
			}
		case "mouse":
			if len(args) < 2 {
				log.Println("Usage: mouse <action> [params...]")
				continue
			}
			if err := a.handleMouseCommand(args[1:]); err != nil {
				log.Printf("Error handling mouse command: %v", err)
			}
		case "key":
			if len(args) < 2 {
				log.Println("Usage: key <action> [params...]")
				continue
			}
			if err := a.handleKeyCommand(args[1:]); err != nil {
				log.Printf("Error handling key command: %v", err)
			}
		case "video":
			if len(args) < 2 {
				log.Println("Usage: video <start|stop>")
				continue
			}
			if err := a.handleVideoCommand(args[1:]); err != nil {
				log.Printf("Error handling video command: %v", err)
			}
		case "record":
			if len(args) < 2 {
				log.Println("Usage: record <start|stop>")
				continue
			}
			if err := a.handleRecordCommand(args[1:]); err != nil {
				log.Printf("Error handling record command: %v", err)
			}
		case "help":
			a.printHelp()
		default:
			log.Printf("Unknown command: %s", command)
			a.printHelp()
		}
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

// testRobotgo runs a series of tests to verify RobotGo functionality
func (a *App) testRobotgo() error {
	log.Println("=================================================================")
	log.Println("🤖 TESTING ROBOTGO FUNCTIONALITY 🤖")
	log.Println("=================================================================")
	log.Println("This test will verify that RobotGo is working correctly.")
	log.Println("You should see the mouse move and draw a square, and text being typed.")
	log.Println("=================================================================")

	// Create a remote controller for testing
	if a.RemoteController == nil {
		log.Println("Creating new RemoteController...")
		a.RemoteController = remote.NewRemoteController(a.PermManager, true) // Force verbose mode for testing
	}

	// Test 1: Get screen size
	log.Println("Test 1: Getting screen size...")
	width, height, err := a.RemoteController.GetScreenSize()
	if err != nil {
		log.Printf("❌ Failed to get screen size: %v", err)
		return fmt.Errorf("failed to get screen size: %w", err)
	}
	log.Printf("✅ Screen size: %dx%d", width, height)

	// Test 2: Get mouse position
	log.Println("Test 2: Getting mouse position...")
	startX, startY, err := a.RemoteController.GetMousePosition()
	if err != nil {
		log.Printf("❌ Failed to get mouse position: %v", err)
		return fmt.Errorf("failed to get mouse position: %w", err)
	}
	log.Printf("✅ Current mouse position: (%d,%d)", startX, startY)

	// Test 3: Move mouse to center of screen
	log.Println("Test 3: Moving mouse to center of screen...")
	centerX := width / 2
	centerY := height / 2
	log.Printf("Attempting to move mouse to (%d,%d)", centerX, centerY)

	err = a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
		Action: remote.MouseMove,
		X:      centerX,
		Y:      centerY,
	})
	if err != nil {
		log.Printf("❌ Failed to move mouse: %v", err)
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	// Verify mouse position after move
	newX, newY, _ := a.RemoteController.GetMousePosition()
	if newX == centerX && newY == centerY {
		log.Printf("✅ Mouse successfully moved to (%d,%d)", newX, newY)
	} else {
		log.Printf("⚠️ Mouse position after move: (%d,%d), expected: (%d,%d)", newX, newY, centerX, centerY)
		log.Println("Mouse movement may not be working correctly.")

		// Ask user if they want to continue
		log.Println("Do you want to continue with the test? (y/n)")
		var input string
		fmt.Scanln(&input)
		if input != "y" && input != "Y" {
			return fmt.Errorf("test aborted by user")
		}
	}

	// Test 4: Draw a square with the mouse
	log.Println("Test 4: Drawing a square with the mouse...")

	// Define square corners (100x100 square around center)
	size := 100
	corners := []struct{ x, y int }{
		{centerX - size/2, centerY - size/2}, // Top-left
		{centerX + size/2, centerY - size/2}, // Top-right
		{centerX + size/2, centerY + size/2}, // Bottom-right
		{centerX - size/2, centerY + size/2}, // Bottom-left
		{centerX - size/2, centerY - size/2}, // Back to top-left
	}

	// Move to first corner
	log.Printf("Moving to first corner: (%d,%d)", corners[0].x, corners[0].y)
	err = a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
		Action: remote.MouseMove,
		X:      corners[0].x,
		Y:      corners[0].y,
	})
	if err != nil {
		log.Printf("❌ Failed to move mouse: %v", err)
		return fmt.Errorf("failed to move mouse: %w", err)
	}

	// Press mouse button down
	log.Println("Pressing mouse button down...")
	err = a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
		Action: remote.MouseDown,
		Button: remote.LeftButton,
	})
	if err != nil {
		log.Printf("❌ Failed to press mouse button: %v", err)
		return fmt.Errorf("failed to press mouse button: %w", err)
	}

	// Draw the square by moving to each corner
	for i := 1; i < len(corners); i++ {
		time.Sleep(500 * time.Millisecond) // Slow down for visibility
		log.Printf("Moving to corner %d: (%d,%d)", i, corners[i].x, corners[i].y)
		err = a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
			Action: remote.MouseMove,
			X:      corners[i].x,
			Y:      corners[i].y,
		})
		if err != nil {
			// Release mouse button before returning error
			a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
				Action: remote.MouseUp,
				Button: remote.LeftButton,
			})
			log.Printf("❌ Failed to move mouse: %v", err)
			return fmt.Errorf("failed to move mouse: %w", err)
		}
	}

	// Release mouse button
	log.Println("Releasing mouse button...")
	err = a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
		Action: remote.MouseUp,
		Button: remote.LeftButton,
	})
	if err != nil {
		log.Printf("❌ Failed to release mouse button: %v", err)
		return fmt.Errorf("failed to release mouse button: %w", err)
	}

	// Test 5: Type some text
	log.Println("Test 5: Testing keyboard input...")
	log.Println("Please open a text editor or click in a text field to see the typing test.")

	// Wait a moment before typing
	log.Println("Waiting 3 seconds before typing...")
	time.Sleep(3 * time.Second)

	// Type a test message
	testText := "RobotGo Test Successful!"
	log.Printf("Typing: \"%s\"", testText)
	err = a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
		Action: remote.KeyType,
		Text:   testText,
	})
	if err != nil {
		log.Printf("❌ Failed to type text: %v", err)
		return fmt.Errorf("failed to type text: %w", err)
	}

	// Move mouse back to original position
	log.Printf("Moving mouse back to original position: (%d,%d)", startX, startY)
	a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
		Action: remote.MouseMove,
		X:      startX,
		Y:      startY,
	})

	log.Println("=================================================================")
	log.Println("✅ All RobotGo tests completed!")
	log.Println("=================================================================")
	return nil
}

// requestPermissionsInteractive requests permissions interactively
func (a *App) requestPermissionsInteractive() error {
	// Create a new permission manager if not already created
	if a.PermManager == nil {
		a.PermManager = permissions.NewManager(a.Config.Verbose)
	}

	fmt.Println("=================================================================")
	fmt.Println("🔒 INTERACTIVE PERMISSION REQUEST 🔒")
	fmt.Println("=================================================================")
	fmt.Println("This will guide you through granting permissions required by the application.")
	fmt.Println("=================================================================")

	// Request screen sharing permission
	fmt.Println("\n1. Screen Sharing Permission")
	fmt.Println("--------------------------")
	screenShareGranted := a.PermManager.RequestPermissionInteractive(permissions.ScreenShare)

	if screenShareGranted {
		fmt.Println("✅ Screen sharing permission granted successfully!")
	} else {
		fmt.Println("⚠️ Screen sharing permission not granted.")
		fmt.Println("Screenshot functionality may not work correctly.")
	}

	// Request remote control permission
	fmt.Println("\n2. Remote Control Permission")
	fmt.Println("--------------------------")
	remoteControlGranted := a.PermManager.RequestPermissionInteractive(permissions.RemoteControl)

	if remoteControlGranted {
		fmt.Println("✅ Remote control permission granted successfully!")
	} else {
		fmt.Println("⚠️ Remote control permission not granted.")
		fmt.Println("Mouse and keyboard control functionality may not work correctly.")
	}

	// Summary
	fmt.Println("\n=================================================================")
	fmt.Println("PERMISSION SUMMARY")
	fmt.Println("=================================================================")
	fmt.Printf("Screen Sharing: %s\n", boolToStatus(screenShareGranted))
	fmt.Printf("Remote Control: %s\n", boolToStatus(remoteControlGranted))
	fmt.Println("=================================================================")

	return nil
}

// boolToStatus converts a boolean to a status string
func boolToStatus(granted bool) string {
	if granted {
		return "✅ Granted"
	}
	return "❌ Not Granted"
}

// initVideoStream initializes the video stream
func (a *App) initVideoStream() error {
	// Convert quality string to video.Quality
	var quality video.Quality
	switch a.Config.VideoQuality {
	case "low":
		quality = video.Low
	case "high":
		quality = video.High
	default:
		quality = video.Medium
	}

	// Create video stream
	a.VideoStream = video.NewVideoStream(quality, a.Config.VideoFPS, a.Config.Verbose)

	// Set callback for frame capture
	a.VideoStream.SetOnFrameCapture(func(frameData []byte) error {
		// Send frame to WebSocket server
		if a.WSClient != nil && a.WSClient.IsConnected() {
			message := map[string]interface{}{
				"type":      MessageTypeVideoFrame,
				"frameData": base64.StdEncoding.EncodeToString(frameData),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			return a.WSClient.SendJSON(message)
		}
		return nil
	})

	// Create video recording directory if needed
	if a.Config.VideoRecording {
		if err := os.MkdirAll(a.Config.VideoRecordingDir, 0755); err != nil {
			return fmt.Errorf("failed to create video recording directory: %w", err)
		}
	}

	// Start video streaming if enabled
	if a.Config.VideoStreaming {
		if err := a.VideoStream.StartStreaming(); err != nil {
			return fmt.Errorf("failed to start video streaming: %w", err)
		}
	}

	// Start video recording if enabled
	if a.Config.VideoRecording {
		if err := a.VideoStream.StartRecording(); err != nil {
			return fmt.Errorf("failed to start video recording: %w", err)
		}
	}

	log.Printf("Video stream initialized with quality %s at %d FPS", a.Config.VideoQuality, a.Config.VideoFPS)
	return nil
}

// startVideoStreaming starts video streaming
func (a *App) startVideoStreaming() error {
	if a.VideoStream == nil {
		if err := a.initVideoStream(); err != nil {
			return err
		}
	}

	if err := a.VideoStream.StartStreaming(); err != nil {
		return fmt.Errorf("failed to start video streaming: %w", err)
	}

	log.Println("Started video streaming")
	return nil
}

// stopVideoStreaming stops video streaming
func (a *App) stopVideoStreaming() {
	if a.VideoStream != nil {
		a.VideoStream.StopStreaming()
		log.Println("Stopped video streaming")
	}
}

// startVideoRecording starts video recording
func (a *App) startVideoRecording() error {
	if a.VideoStream == nil {
		if err := a.initVideoStream(); err != nil {
			return err
		}
	}

	if err := a.VideoStream.StartRecording(); err != nil {
		return fmt.Errorf("failed to start video recording: %w", err)
	}

	// Send recording status update to the server
	if a.WSClient != nil && a.WSClient.IsConnected() {
		statusMsg := map[string]interface{}{
			"type":      MessageTypeScreenRecordingStatus,
			"status":    "recording",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		if err := a.WSClient.SendJSON(statusMsg); err != nil {
			log.Printf("Failed to send recording status update: %v", err)
		}
	}

	log.Println("Started video recording")
	return nil
}

// stopVideoRecording stops video recording and saves the recording
func (a *App) stopVideoRecording() error {
	if a.VideoStream == nil {
		return fmt.Errorf("video stream not initialized")
	}

	frames, err := a.VideoStream.StopRecording()
	if err != nil {
		return fmt.Errorf("failed to stop video recording: %w", err)
	}

	log.Printf("Stopped video recording, captured %d frames", len(frames))

	// Send recording status update to the server
	if a.WSClient != nil && a.WSClient.IsConnected() {
		statusMsg := map[string]interface{}{
			"type":      MessageTypeScreenRecordingStatus,
			"status":    "stopped",
			"frames":    len(frames),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		if err := a.WSClient.SendJSON(statusMsg); err != nil {
			log.Printf("Failed to send recording status update: %v", err)
		}
	}

	// Save recording as images
	timestamp := time.Now().Format("20060102-150405")
	recordingDir := filepath.Join(a.Config.VideoRecordingDir, timestamp)
	if err := os.MkdirAll(recordingDir, 0755); err != nil {
		return fmt.Errorf("failed to create recording directory: %w", err)
	}

	if err := a.VideoStream.SaveRecordingAsImages(recordingDir, "frame"); err != nil {
		return fmt.Errorf("failed to save recording: %w", err)
	}

	// Send saved recording notification to the server
	if a.WSClient != nil && a.WSClient.IsConnected() {
		savedMsg := map[string]interface{}{
			"type":        MessageTypeScreenRecordingSaved,
			"directory":   recordingDir,
			"frameCount":  len(frames),
			"timestamp":   time.Now().Format(time.RFC3339),
			"recordingId": timestamp,
		}
		if err := a.WSClient.SendJSON(savedMsg); err != nil {
			log.Printf("Failed to send recording saved notification: %v", err)
		}
	}

	log.Printf("Saved recording to %s", recordingDir)
	return nil
}

// handleVideoCommand handles video streaming commands
func (a *App) handleVideoCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no video command specified")
	}

	switch args[0] {
	case "start":
		return a.startVideoStreaming()
	case "stop":
		a.stopVideoStreaming()
		return nil
	case "status":
		if a.VideoStream == nil {
			log.Println("Video stream not initialized")
		} else {
			log.Printf("Video streaming: %v", a.VideoStream.IsStreaming())
			log.Printf("Video recording: %v", a.VideoStream.IsRecording())
		}
		return nil
	default:
		return fmt.Errorf("unknown video command: %s", args[0])
	}
}

// handleRecordCommand handles video recording commands
func (a *App) handleRecordCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no record command specified")
	}

	switch args[0] {
	case "start":
		return a.startVideoRecording()
	case "stop":
		return a.stopVideoRecording()
	case "status":
		return a.getRecordingStatus()
	default:
		return fmt.Errorf("unknown record command: %s", args[0])
	}
}

// printHelp prints the help message
func (a *App) printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  screenshot [quality]       - Take a screenshot (quality: low, medium, high)")
	fmt.Println("  region <x> <y> <w> <h>     - Take a screenshot of a specific region")
	fmt.Println("  auto [off]                 - Start/stop automatic screenshots")
	fmt.Println("  mouse <action> [params...] - Perform a mouse action")
	fmt.Println("  key <action> [params...]   - Perform a keyboard action")
	fmt.Println("  video <start|stop|status>  - Control video streaming")
	fmt.Println("  record <start|stop|status> - Control video recording")
	fmt.Println("  help                       - Show this help message")
	fmt.Println("  exit, quit                 - Exit the application")
}

// handleMouseCommand handles mouse commands
func (a *App) handleMouseCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no mouse command specified")
	}

	action := args[0]
	switch action {
	case "move":
		if len(args) < 3 {
			return fmt.Errorf("usage: mouse move <x> <y>")
		}
		x, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid x coordinate: %w", err)
		}
		y, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid y coordinate: %w", err)
		}
		return a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
			Action: remote.MouseMove,
			X:      x,
			Y:      y,
		})
	case "click":
		button := remote.LeftButton
		if len(args) > 1 {
			switch args[1] {
			case "right":
				button = remote.RightButton
			case "middle":
				button = remote.MiddleButton
			}
		}
		return a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
			Action: remote.MouseClick,
			Button: button,
		})
	case "down":
		button := remote.LeftButton
		if len(args) > 1 {
			switch args[1] {
			case "right":
				button = remote.RightButton
			case "middle":
				button = remote.MiddleButton
			}
		}
		return a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
			Action: remote.MouseDown,
			Button: button,
		})
	case "up":
		button := remote.LeftButton
		if len(args) > 1 {
			switch args[1] {
			case "right":
				button = remote.RightButton
			case "middle":
				button = remote.MiddleButton
			}
		}
		return a.RemoteController.ExecuteMouseEvent(remote.MouseEvent{
			Action: remote.MouseUp,
			Button: button,
		})
	case "position":
		x, y, err := a.RemoteController.GetMousePosition()
		if err != nil {
			return fmt.Errorf("failed to get mouse position: %w", err)
		}
		log.Printf("Mouse position: (%d,%d)", x, y)
		return nil
	default:
		return fmt.Errorf("unknown mouse command: %s", action)
	}
}

// handleKeyCommand handles keyboard commands
func (a *App) handleKeyCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no key command specified")
	}

	action := args[0]
	switch action {
	case "press":
		if len(args) < 2 {
			return fmt.Errorf("usage: key press <key>")
		}
		return a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
			Action: remote.KeyPress,
			Key:    args[1],
		})
	case "down":
		if len(args) < 2 {
			return fmt.Errorf("usage: key down <key>")
		}
		return a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
			Action: remote.KeyDown,
			Key:    args[1],
		})
	case "up":
		if len(args) < 2 {
			return fmt.Errorf("usage: key up <key>")
		}
		return a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
			Action: remote.KeyUp,
			Key:    args[1],
		})
	case "type":
		if len(args) < 2 {
			return fmt.Errorf("usage: key type <text>")
		}
		text := strings.Join(args[1:], " ")
		return a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
			Action: remote.KeyType,
			Text:   text,
		})
	case "combo":
		if len(args) < 2 {
			return fmt.Errorf("usage: key combo <key1> <key2> ...")
		}
		return a.RemoteController.ExecuteKeyboardEvent(remote.KeyboardEvent{
			Action: remote.KeyCombination,
			Keys:   args[1:],
		})
	default:
		return fmt.Errorf("unknown key command: %s", action)
	}
}

// getRecordingStatus gets the current recording status and sends it to the server
func (a *App) getRecordingStatus() error {
	if a.VideoStream == nil {
		// Send status that no video stream is initialized
		if a.WSClient != nil && a.WSClient.IsConnected() {
			statusMsg := map[string]interface{}{
				"type":      MessageTypeScreenRecordingStatus,
				"status":    "not_initialized",
				"timestamp": time.Now().Format(time.RFC3339),
			}
			return a.WSClient.SendJSON(statusMsg)
		}
		return nil
	}

	// Get recording status
	isRecording := a.VideoStream.IsRecording()
	isStreaming := a.VideoStream.IsStreaming()
	frameCount := 0
	if isRecording {
		frameCount = a.VideoStream.GetFrameCount()
	}

	// Send status to the server
	if a.WSClient != nil && a.WSClient.IsConnected() {
		status := "idle"
		if isRecording {
			status = "recording"
		} else if isStreaming {
			status = "streaming"
		}

		statusMsg := map[string]interface{}{
			"type":        MessageTypeScreenRecordingStatus,
			"status":      status,
			"isRecording": isRecording,
			"isStreaming": isStreaming,
			"frameCount":  frameCount,
			"timestamp":   time.Now().Format(time.RFC3339),
		}
		return a.WSClient.SendJSON(statusMsg)
	}

	return nil
}
