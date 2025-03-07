package permissions

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// For testing purposes, we can replace these with mocks
var (
	execCommand  = exec.Command
	execLookPath = exec.LookPath
	osRemove     = os.Remove
)

// requestMacOSScreenSharePermission requests screen sharing permission on macOS
func requestMacOSScreenSharePermission() (PermissionStatus, error) {
	if runtime.GOOS != "darwin" {
		return Unknown, fmt.Errorf("macOS screen sharing permission check only available on macOS")
	}

	// First, try to check if we already have screen recording permission
	// by attempting to capture a small screenshot
	if checkMacOSScreenCapturePermission() {
		return Granted, nil
	}

	// Request permission by showing a dialog to the user
	fmt.Println("=== Screen Recording Permission Required ===")
	fmt.Println("This application needs screen recording permission to function properly.")
	fmt.Println("Please follow these steps:")
	fmt.Println("1. Click 'OK' when the system dialog appears")
	fmt.Println("2. Check the box next to this application in the Screen Recording section")
	fmt.Println("3. Click 'Quit & Reopen' if prompted, or manually close and reopen the app")
	fmt.Println("=======================================")

	// Open the System Preferences to the Screen Recording section
	openCmd := execCommand("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
	if err := openCmd.Run(); err != nil {
		return Denied, fmt.Errorf("failed to open system preferences: %w", err)
	}

	// Ask the user if they want to retry the permission check
	fmt.Println("\nAfter granting permission in System Preferences:")
	fmt.Println("1. Press 'r' to retry the permission check")
	fmt.Println("2. Press 'q' to quit the application")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nEnter your choice (r/q): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "r" || input == "R" {
			fmt.Println("Retrying permission check...")
			// Wait a moment before retrying
			time.Sleep(500 * time.Millisecond)

			if checkMacOSScreenCapturePermission() {
				fmt.Println("Screen recording permission granted successfully!")
				return Granted, nil
			}

			fmt.Println("Permission not granted yet. Please make sure you've completed all steps.")
			fmt.Println("If you've already granted permission, you may need to quit and restart the application.")
		} else if input == "q" || input == "Q" {
			fmt.Println("Exiting application. Please restart after granting permission.")
			return Requested, nil
		} else {
			fmt.Println("Invalid input. Please enter 'r' to retry or 'q' to quit.")
		}
	}
}

// checkMacOSScreenCapturePermission checks if screen capture permission is granted on macOS
func checkMacOSScreenCapturePermission() bool {
	// Try to capture a 1x1 pixel screenshot as a test
	// This will trigger the permission check if not already granted
	cmd := execCommand("screencapture", "-x", "-t", "png", "-R", "0,0,1,1", "/tmp/permission_test.png")
	err := cmd.Run()

	// Clean up the test file
	osRemove("/tmp/permission_test.png")

	// If the command succeeded, we have permission
	return err == nil
}

// requestWindowsScreenSharePermission requests screen sharing permission on Windows
func requestWindowsScreenSharePermission() (PermissionStatus, error) {
	if runtime.GOOS != "windows" {
		return Unknown, fmt.Errorf("Windows screen sharing permission check only available on Windows")
	}

	// On Windows 10 and later, we can check and request screen capture permissions
	// through the Settings app

	fmt.Println("Screen sharing permission is required.")
	fmt.Println("Please ensure screen capture is enabled in Windows Settings > Privacy > Camera")

	// Open Windows Settings to the relevant privacy section
	cmd := execCommand("start", "ms-settings:privacy-screenrecording")
	if err := cmd.Run(); err != nil {
		return Denied, fmt.Errorf("failed to open Windows settings: %w", err)
	}

	fmt.Println("After ensuring permissions are granted, please press Enter to continue...")
	fmt.Scanln() // Wait for user to press Enter

	// For simplicity, we'll assume the user granted permission
	// In a production app, you would implement a proper check here
	return Granted, nil
}

// requestLinuxScreenSharePermission requests screen sharing permission on Linux
func requestLinuxScreenSharePermission() (PermissionStatus, error) {
	if runtime.GOOS != "linux" {
		return Unknown, fmt.Errorf("Linux screen sharing permission check only available on Linux")
	}

	// On Linux, screen sharing permissions depend on the desktop environment
	// For this example, we'll provide a generic approach

	fmt.Println("Screen sharing permission is required.")

	// Check if we have XDG-based desktop environment
	_, err := execLookPath("xdg-open")
	if err == nil {
		fmt.Println("Please ensure your desktop environment allows screen sharing.")

		// Some desktop environments might have specific settings
		// Try to open the relevant settings if possible
		openCmd := execCommand("xdg-open", "settings://privacy")
		_ = openCmd.Run() // Ignore errors as this might not work on all desktop environments
	} else {
		fmt.Println("Please ensure your window manager or desktop environment allows screen sharing.")
	}

	fmt.Println("After ensuring permissions are granted, please press Enter to continue...")
	fmt.Scanln() // Wait for user to press Enter

	// For simplicity, we'll assume the user granted permission
	// In a production app, you would implement a proper check here
	return Granted, nil
}
