package main

import (
	"log"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	log.Println("=================================================================")
	log.Println("🖱️ MOUSE CONTROL TEST 🖱️")
	log.Println("=================================================================")
	log.Println("This test will try different methods to control the mouse.")
	log.Println("=================================================================")

	// Get screen size
	width, height := robotgo.GetScreenSize()
	log.Printf("Screen size: %dx%d", width, height)

	// Get current mouse position
	startX, startY := robotgo.GetMousePos()
	log.Printf("Starting mouse position: (%d,%d)", startX, startY)

	// Calculate center of screen
	centerX := width / 2
	centerY := height / 2
	log.Printf("Screen center: (%d,%d)", centerX, centerY)

	// Test 1: Basic Move
	log.Println("\nTest 1: Basic Move")
	log.Printf("Moving mouse to (%d,%d) using robotgo.Move", centerX, centerY)
	robotgo.Move(centerX, centerY)
	time.Sleep(500 * time.Millisecond)

	// Check position
	x, y := robotgo.GetMousePos()
	log.Printf("Position after move: (%d,%d)", x, y)
	if x == centerX && y == centerY {
		log.Println("✅ Basic Move successful")
	} else {
		log.Println("❌ Basic Move failed")
	}

	// Test 2: MoveSmooth
	log.Println("\nTest 2: MoveSmooth")
	// Move to a different position first
	newX := centerX + 100
	newY := centerY + 100
	log.Printf("Moving mouse to (%d,%d) using robotgo.MoveMouseSmooth", newX, newY)
	robotgo.MoveMouseSmooth(newX, newY, 1.0, 1.0)
	time.Sleep(500 * time.Millisecond)

	// Check position
	x, y = robotgo.GetMousePos()
	log.Printf("Position after move: (%d,%d)", x, y)
	if x == newX && y == newY {
		log.Println("✅ MoveSmooth successful")
	} else {
		log.Println("❌ MoveSmooth failed")
	}

	// Test 3: MoveRelative
	log.Println("\nTest 3: MoveRelative")
	log.Printf("Moving mouse relatively by (-100, -100) using robotgo.MoveRelative")
	robotgo.MoveRelative(-100, -100)
	time.Sleep(500 * time.Millisecond)

	// Check position
	x, y = robotgo.GetMousePos()
	expectedX := newX - 100
	expectedY := newY - 100
	log.Printf("Position after move: (%d,%d), expected: (%d,%d)", x, y, expectedX, expectedY)
	if x == expectedX && y == expectedY {
		log.Println("✅ MoveRelative successful")
	} else {
		log.Println("❌ MoveRelative failed")
	}

	// Test 4: DragSmooth
	log.Println("\nTest 4: DragSmooth")
	dragX := centerX - 100
	dragY := centerY - 100
	log.Printf("Dragging mouse to (%d,%d) using robotgo.DragSmooth", dragX, dragY)
	robotgo.DragSmooth(dragX, dragY)
	time.Sleep(500 * time.Millisecond)

	// Check position
	x, y = robotgo.GetMousePos()
	log.Printf("Position after drag: (%d,%d)", x, y)
	if x == dragX && y == dragY {
		log.Println("✅ DragSmooth successful")
	} else {
		log.Println("❌ DragSmooth failed")
	}

	// Test 5: Toggle and Move
	log.Println("\nTest 5: Toggle and Move")
	log.Println("Pressing mouse button down, moving, and releasing")

	// Move to start position
	startToggleX := centerX - 50
	startToggleY := centerY - 50
	log.Printf("Moving to start position: (%d,%d)", startToggleX, startToggleY)
	robotgo.Move(startToggleX, startToggleY)
	time.Sleep(500 * time.Millisecond)

	// Press mouse button
	log.Println("Pressing mouse button down")
	robotgo.Toggle("left")
	time.Sleep(500 * time.Millisecond)

	// Move while pressed
	endToggleX := centerX + 50
	endToggleY := centerY + 50
	log.Printf("Moving to end position: (%d,%d)", endToggleX, endToggleY)
	robotgo.Move(endToggleX, endToggleY)
	time.Sleep(500 * time.Millisecond)

	// Release mouse button
	log.Println("Releasing mouse button")
	robotgo.Toggle("left", "up")
	time.Sleep(500 * time.Millisecond)

	// Check position
	x, y = robotgo.GetMousePos()
	log.Printf("Final position: (%d,%d)", x, y)
	if x == endToggleX && y == endToggleY {
		log.Println("✅ Toggle and Move successful")
	} else {
		log.Println("❌ Toggle and Move failed")
	}

	// Move back to original position
	log.Printf("\nMoving back to original position: (%d,%d)", startX, startY)
	robotgo.Move(startX, startY)

	log.Println("=================================================================")
	log.Println("Mouse control test completed")
	log.Println("=================================================================")
}
