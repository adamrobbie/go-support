package main

import (
	"log"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	log.Println("Starting RobotGo test application...")

	// Test 1: Get screen size
	log.Println("Test 1: Getting screen size...")
	width, height := robotgo.GetScreenSize()
	log.Printf("Screen size: %dx%d", width, height)

	// Test 2: Get mouse position
	log.Println("Test 2: Getting mouse position...")
	x, y := robotgo.GetMousePos()
	log.Printf("Current mouse position: (%d,%d)", x, y)

	// Test 3: Move mouse to center of screen
	log.Println("Test 3: Moving mouse to center of screen...")
	centerX := width / 2
	centerY := height / 2
	robotgo.MoveMouseSmooth(centerX, centerY, 1.0, 1.0)
	log.Printf("Moved mouse to (%d,%d)", centerX, centerY)

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
	robotgo.MoveMouseSmooth(corners[0].x, corners[0].y, 1.0, 1.0)

	// Press mouse button down
	robotgo.Toggle("left")

	// Draw the square by moving to each corner
	for i := 1; i < len(corners); i++ {
		time.Sleep(500 * time.Millisecond) // Slow down for visibility
		robotgo.MoveMouseSmooth(corners[i].x, corners[i].y, 1.0, 1.0)
		log.Printf("Moved to corner %d: (%d,%d)", i, corners[i].x, corners[i].y)
	}

	// Release mouse button
	robotgo.Toggle("left", "up")

	// Test 5: Type some text
	log.Println("Test 5: Testing keyboard input...")

	// Wait a moment before typing
	time.Sleep(1 * time.Second)

	// Type a test message
	robotgo.TypeStr("RobotGo Test Successful!")

	// Test 6: Take a screenshot
	log.Println("Test 6: Taking a screenshot...")

	// Capture the screen
	robotgo.SaveCapture("robotgo-test-screenshot.png")
	log.Println("Screenshot saved to robotgo-test-screenshot.png")

	log.Println("All RobotGo tests completed successfully!")
}
