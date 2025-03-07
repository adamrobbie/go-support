package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		log.Fatalf("This test is for macOS only. Current OS: %s", runtime.GOOS)
	}

	log.Println("=================================================================")
	log.Println("🍎 MACOS ACCESSIBILITY TEST 🍎")
	log.Println("=================================================================")
	log.Println("This test will check for macOS-specific accessibility issues.")
	log.Println("=================================================================")

	// Test 1: Check Accessibility permissions using AppleScript
	log.Println("\nTest 1: Check Accessibility permissions using AppleScript")
	cmd := exec.Command("osascript", "-e", `tell application "System Events" to keystroke ""`)
	err := cmd.Run()
	if err != nil {
		log.Printf("❌ AppleScript test failed: %v", err)
		log.Println("This indicates that accessibility permissions are not granted.")
		log.Println("Please go to System Preferences > Security & Privacy > Privacy > Accessibility")
		log.Println("and make sure this application is allowed.")
	} else {
		log.Println("✅ AppleScript test passed")
	}

	// Test 2: Check if we can get mouse position
	log.Println("\nTest 2: Check if we can get mouse position")
	x, y := robotgo.GetMousePos()
	log.Printf("Current mouse position: (%d,%d)", x, y)
	if x == 0 && y == 0 {
		log.Println("⚠️ Mouse position is (0,0), which might indicate a problem")
	} else {
		log.Println("✅ Successfully got mouse position")
	}

	// Test 3: Try to move mouse with CGEventPost (lower level)
	log.Println("\nTest 3: Try to move mouse with CGEventPost")
	log.Println("Attempting to move mouse to center of screen...")

	// Get screen size
	width, height := robotgo.GetScreenSize()
	centerX := width / 2
	centerY := height / 2

	// Try to move mouse using CGEventPost (what robotgo uses internally)
	cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to set mouse position to {%d, %d}`, centerX, centerY))
	err = cmd.Run()
	if err != nil {
		log.Printf("❌ AppleScript mouse move failed: %v", err)
	} else {
		log.Println("✅ AppleScript mouse move command executed")
	}

	// Check if mouse actually moved
	time.Sleep(500 * time.Millisecond)
	newX, newY := robotgo.GetMousePos()
	log.Printf("Mouse position after move: (%d,%d)", newX, newY)
	if newX == centerX && newY == centerY {
		log.Println("✅ Mouse successfully moved to center")
	} else {
		log.Println("❌ Mouse did not move to expected position")
	}

	// Test 4: Check for macOS version-specific issues
	log.Println("\nTest 4: Check for macOS version-specific issues")
	cmd = exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("❌ Failed to get macOS version: %v", err)
	} else {
		version := string(output)
		log.Printf("macOS version: %s", version)

		// Check for known problematic versions
		if version >= "10.15" { // Catalina and newer
			log.Println("⚠️ macOS Catalina and newer have stricter security requirements")
			log.Println("Make sure you've granted permissions in System Preferences")
		}

		if version >= "11.0" { // Big Sur and newer
			log.Println("⚠️ macOS Big Sur and newer may require additional permissions")
			log.Println("You might need to grant Full Disk Access in addition to Accessibility")
		}
	}

	// Test 5: Try alternative mouse movement method
	log.Println("\nTest 5: Try alternative mouse movement method")
	altX := centerX + 100
	altY := centerY + 100
	log.Printf("Moving mouse to (%d,%d) using alternative method...", altX, altY)

	// Use cliclick if available (a command-line tool for mouse control)
	cmd = exec.Command("which", "cliclick")
	if err := cmd.Run(); err == nil {
		// cliclick is installed
		cmd = exec.Command("cliclick", fmt.Sprintf("m:%d,%d", altX, altY))
		if err := cmd.Run(); err != nil {
			log.Printf("❌ cliclick failed: %v", err)
		} else {
			log.Println("✅ cliclick command executed")
			time.Sleep(500 * time.Millisecond)
			x, y = robotgo.GetMousePos()
			log.Printf("Mouse position after cliclick: (%d,%d)", x, y)
		}
	} else {
		log.Println("⚠️ cliclick not installed, skipping alternative method")
		log.Println("You can install it with: brew install cliclick")
	}

	log.Println("=================================================================")
	log.Println("macOS accessibility test completed")
	log.Println("=================================================================")

	// Recommendations
	log.Println("\nRecommendations:")
	log.Println("1. Make sure the application has Accessibility permissions")
	log.Println("2. Try running the application with sudo (not recommended for production)")
	log.Println("3. Try installing and using cliclick as an alternative")
	log.Println("4. Consider using the mouse-test tool to try different movement methods")
}
