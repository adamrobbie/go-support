//go:build darwin
// +build darwin

package remote

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// macOSMoveMouse moves the mouse using AppleScript as a fallback method
// This is used when RobotGo's native methods fail
func macOSMoveMouse(x, y int, verbose bool) error {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("macOSMoveMouse is only supported on macOS")
	}

	if verbose {
		log.Printf("Using AppleScript fallback to move mouse to (%d,%d)", x, y)
	}

	// Use AppleScript to move the mouse
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to set mouse position to {%d, %d}`, x, y))
	err := cmd.Run()
	if err != nil {
		if verbose {
			log.Printf("AppleScript mouse move failed: %v", err)
		}
		return err
	}

	if verbose {
		log.Printf("AppleScript mouse move executed")
	}

	return nil
}

// macOSClickMouse clicks the mouse using AppleScript as a fallback method
func macOSClickMouse(button string, double bool, verbose bool) error {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("macOSClickMouse is only supported on macOS")
	}

	buttonNum := 1 // Default to left button
	if button == "right" {
		buttonNum = 2
	} else if button == "center" {
		buttonNum = 3
	}

	if verbose {
		log.Printf("Using AppleScript fallback to click mouse button %s (button %d)", button, buttonNum)
	}

	// Construct the AppleScript command
	script := fmt.Sprintf(`tell application "System Events" to click button %d of (get mouse)`, buttonNum)
	if double {
		script = fmt.Sprintf(`tell application "System Events" to click button %d of (get mouse) 2 times`, buttonNum)
	}

	// Execute the AppleScript
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		if verbose {
			log.Printf("AppleScript mouse click failed: %v", err)
		}
		return err
	}

	if verbose {
		log.Printf("AppleScript mouse click executed")
	}

	return nil
}

// macOSToggleMouse presses or releases a mouse button using AppleScript as a fallback method
func macOSToggleMouse(button string, direction string, verbose bool) error {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("macOSToggleMouse is only supported on macOS")
	}

	buttonNum := 1 // Default to left button
	if button == "right" {
		buttonNum = 2
	} else if button == "center" {
		buttonNum = 3
	}

	action := "down"
	if direction == "up" {
		action = "up"
	}

	if verbose {
		log.Printf("Using AppleScript fallback to toggle mouse button %s %s (button %d)", button, action, buttonNum)
	}

	// Construct the AppleScript command
	script := ""
	if action == "down" {
		script = fmt.Sprintf(`tell application "System Events" to mouse button %d down`, buttonNum)
	} else {
		script = fmt.Sprintf(`tell application "System Events" to mouse button %d up`, buttonNum)
	}

	// Execute the AppleScript
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		if verbose {
			log.Printf("AppleScript mouse toggle failed: %v", err)
		}
		return err
	}

	if verbose {
		log.Printf("AppleScript mouse toggle executed")
	}

	return nil
}

// macOSTypeText types text using AppleScript as a fallback method
func macOSTypeText(text string, verbose bool) error {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("macOSTypeText is only supported on macOS")
	}

	if verbose {
		log.Printf("Using AppleScript fallback to type text: %s", text)
	}

	// Use AppleScript to type text
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, text))
	err := cmd.Run()
	if err != nil {
		if verbose {
			log.Printf("AppleScript text typing failed: %v", err)
		}
		return err
	}

	if verbose {
		log.Printf("AppleScript text typing executed")
	}

	return nil
}

// macOSKeyTap presses a key using AppleScript as a fallback method
func macOSKeyTap(key string, verbose bool) error {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("macOSKeyTap is only supported on macOS")
	}

	if verbose {
		log.Printf("Using AppleScript fallback to tap key: %s", key)
	}

	// Map common keys to AppleScript key codes
	keyMap := map[string]string{
		"enter":     "return",
		"return":    "return",
		"tab":       "tab",
		"space":     "space",
		"backspace": "delete",
		"delete":    "delete",
		"escape":    "escape",
		"up":        "up arrow",
		"down":      "down arrow",
		"left":      "left arrow",
		"right":     "right arrow",
		"home":      "home",
		"end":       "end",
		"pageup":    "page up",
		"pagedown":  "page down",
	}

	// Get the AppleScript key name
	keyName, ok := keyMap[key]
	if !ok {
		keyName = key // Use the key as is if not in the map
	}

	// Use AppleScript to press the key
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "System Events" to key code "%s"`, keyName))
	err := cmd.Run()
	if err != nil {
		if verbose {
			log.Printf("AppleScript key tap failed: %v", err)
		}
		return err
	}

	if verbose {
		log.Printf("AppleScript key tap executed")
	}

	return nil
}
