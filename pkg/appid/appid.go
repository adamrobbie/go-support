package appid

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// AppName is the name of the application
	AppName = "Go Support"

	// AppID is the unique identifier for the application
	AppID = "com.adamrobbie.go-support"

	// AppVersion is the version of the application
	AppVersion = "1.0.0"

	// AppDescription is a short description of the application
	AppDescription = "A cross-platform WebSocket client with screen sharing capabilities"
)

// SetupAppIdentifier configures the application identifier for the current platform
func SetupAppIdentifier() error {
	switch runtime.GOOS {
	case "darwin":
		return setupMacOSIdentifier()
	case "windows":
		return setupWindowsIdentifier()
	case "linux":
		return setupLinuxIdentifier()
	default:
		return fmt.Errorf("unsupported platform for application identification: %s", runtime.GOOS)
	}
}

// setupMacOSIdentifier sets up the application identifier for macOS
func setupMacOSIdentifier() error {
	// Print application information
	fmt.Printf("Application: %s (ID: %s)\n", AppName, AppID)
	fmt.Printf("Version: %s\n", AppVersion)

	// Set up macOS app bundle
	if err := SetupMacOSAppBundle(); err != nil {
		fmt.Printf("Warning: Failed to set up macOS app bundle: %v\n", err)
		// Continue even if this fails
	}

	return nil
}

// setupWindowsIdentifier sets up the application identifier for Windows
func setupWindowsIdentifier() error {
	// On Windows, we can set the Application User Model ID (AUMID)
	// This is typically done for GUI applications, but we'll include it for completeness
	fmt.Printf("Application: %s (ID: %s)\n", AppName, AppID)
	fmt.Printf("Version: %s\n", AppVersion)

	return nil
}

// setupLinuxIdentifier sets up the application identifier for Linux
func setupLinuxIdentifier() error {
	// On Linux, we can create a desktop entry file
	fmt.Printf("Application: %s (ID: %s)\n", AppName, AppID)
	fmt.Printf("Version: %s\n", AppVersion)

	// Create a desktop entry file in the user's local applications directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(desktopDir, 0755); err != nil {
		return fmt.Errorf("failed to create desktop directory: %w", err)
	}

	desktopFile := filepath.Join(desktopDir, "go-support.desktop")
	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Comment=%s
Exec=%s
Terminal=true
Categories=Network;Utility;
X-GNOME-UsesNotifications=true
`, AppName, AppDescription, os.Args[0])

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	return nil
}
