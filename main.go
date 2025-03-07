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

func main() {
	// Set up application identifier
	if err := appid.SetupAppIdentifier(); err != nil {
		log.Printf("Warning: Failed to set up application identifier: %v", err)
	}

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Get WebSocket URL from environment variable
	wsURL := os.Getenv("WEBSOCKET_URL")
	if wsURL == "" {
		log.Fatal("WEBSOCKET_URL environment variable not set")
	}

	// Parse command line flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	skipPermissions := flag.Bool("skip-permissions", false, "Skip permission checks")
	flag.Parse()

	if *verbose {
		log.Println("Verbose mode enabled")
	}

	// Initialize permission manager
	permManager := permissions.NewManager()

	// Request screen sharing permission if not skipped
	if !*skipPermissions {
		log.Println("Checking screen sharing permission...")
		status, err := permManager.RequestPermission(permissions.ScreenShare)
		if err != nil {
			log.Printf("Warning: Failed to request screen sharing permission: %v", err)
			log.Println("Continuing without screen sharing permission...")
		} else {
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
				return
			default:
				log.Println("Screen sharing permission status unknown")
				log.Println("Continuing anyway...")
			}
		}
	} else {
		log.Println("Skipping permission checks as requested")
	}

	// Log that we're connecting to the WebSocket
	log.Printf("Connecting to WebSocket at %s", wsURL)

	// Connect to WebSocket
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer c.Close()

	log.Println("Connected to WebSocket server")

	// Set up channel to handle interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Set up channel for receiving messages
	done := make(chan struct{})

	// Start goroutine to read messages from WebSocket
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("Error reading from WebSocket:", err)
				return
			}
			fmt.Printf("Received: %s\n", message)
		}
	}()

	// Send periodic ping messages
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Println("Error sending ping:", err)
				return
			}
			if *verbose {
				log.Println("Sent ping")
			}
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")

			// Cleanly close the connection by sending a close message
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing WebSocket:", err)
				return
			}

			// Wait for the server to close the connection
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
