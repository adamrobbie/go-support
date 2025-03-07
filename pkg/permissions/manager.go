package permissions

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// PermissionType represents different types of permissions
type PermissionType string

// String returns the string representation of PermissionType
func (p PermissionType) String() string {
	return string(p)
}

const (
	// ScreenShare permission for screen sharing
	ScreenShare PermissionType = "screen_share"
	// RemoteControl permission for keyboard and mouse control
	RemoteControl PermissionType = "remote_control"
	// Add more permission types as needed
)

// PermissionStatus represents the status of a permission
type PermissionStatus int

// String returns the string representation of PermissionStatus
func (p PermissionStatus) String() string {
	switch p {
	case Unknown:
		return "Unknown"
	case Granted:
		return "Granted"
	case Denied:
		return "Denied"
	case Requested:
		return "Requested"
	default:
		return fmt.Sprintf("Unknown Status: %d", p)
	}
}

const (
	// Unknown permission status is not determined
	Unknown PermissionStatus = iota
	// Granted permission is granted
	Granted
	// Denied permission is denied
	Denied
	// Requested permission is requested but waiting for user response
	Requested
)

// Manager handles permission requests and checks
type Manager interface {
	// RequestPermission requests a specific permission
	RequestPermission(permType PermissionType) (PermissionStatus, error)

	// CheckPermission checks if a specific permission is granted
	CheckPermission(permType PermissionType) (PermissionStatus, error)

	// EnsurePermission checks if a permission is granted and requests it if not
	// It returns true if the permission is granted, false otherwise
	EnsurePermission(permType PermissionType) (bool, error)

	// RequestPermissionInteractive requests a permission with an interactive flow
	// It returns true if the permission was granted, false otherwise
	RequestPermissionInteractive(permType PermissionType) bool
}

// DefaultManager is the default implementation of Manager
type DefaultManager struct {
	permissions map[PermissionType]PermissionStatus
	verbose     bool
}

// NewManager creates a new permission manager
func NewManager(verbose bool) Manager {
	return &DefaultManager{
		permissions: make(map[PermissionType]PermissionStatus),
		verbose:     verbose,
	}
}

// RequestPermission implements the Manager interface
func (m *DefaultManager) RequestPermission(permType PermissionType) (PermissionStatus, error) {
	// Check if we already have the permission
	status, err := m.CheckPermission(permType)
	if err == nil && status == Granted {
		return Granted, nil
	}

	// Handle different permission types
	switch permType {
	case ScreenShare:
		return m.requestScreenSharePermission()
	case RemoteControl:
		return m.requestRemoteControlPermission()
	default:
		return Unknown, fmt.Errorf("unsupported permission type: %s", permType)
	}
}

// CheckPermission implements the Manager interface
func (m *DefaultManager) CheckPermission(permType PermissionType) (PermissionStatus, error) {
	// First check if we have a cached status
	status, exists := m.permissions[permType]
	if exists {
		return status, nil
	}

	// If not cached, check the actual permission status
	switch permType {
	case ScreenShare:
		return m.checkScreenSharePermission()
	case RemoteControl:
		return m.checkRemoteControlPermission()
	default:
		return Unknown, nil
	}
}

// EnsurePermission implements the Manager interface
func (m *DefaultManager) EnsurePermission(permType PermissionType) (bool, error) {
	// First check if we already have the permission
	status, err := m.CheckPermission(permType)
	if err != nil {
		return false, err
	}

	if status == Granted {
		return true, nil
	}

	// If not granted, request the permission
	status, err = m.RequestPermission(permType)
	if err != nil {
		return false, err
	}

	// Return true only if the permission is granted
	return status == Granted, nil
}

// requestScreenSharePermission requests screen sharing permission based on the platform
func (m *DefaultManager) requestScreenSharePermission() (PermissionStatus, error) {
	var status PermissionStatus
	var err error

	switch runtime.GOOS {
	case "darwin":
		status, err = m.requestMacOSScreenSharePermission()
	case "windows":
		status, err = m.requestWindowsScreenSharePermission()
	case "linux":
		status, err = m.requestLinuxScreenSharePermission()
	default:
		return Unknown, errors.New("unsupported platform for screen sharing")
	}

	if err == nil {
		m.permissions[ScreenShare] = status
	}
	return status, err
}

// checkScreenSharePermission checks screen sharing permission based on the platform
func (m *DefaultManager) checkScreenSharePermission() (PermissionStatus, error) {
	var status PermissionStatus
	var err error

	switch runtime.GOOS {
	case "darwin":
		status, err = m.checkMacOSScreenSharePermission()
	case "windows":
		status, err = m.checkWindowsScreenSharePermission()
	case "linux":
		status, err = m.checkLinuxScreenSharePermission()
	default:
		return Unknown, errors.New("unsupported platform for screen sharing")
	}

	if err == nil {
		m.permissions[ScreenShare] = status
	}
	return status, err
}

// requestRemoteControlPermission requests remote control permission based on the platform
func (m *DefaultManager) requestRemoteControlPermission() (PermissionStatus, error) {
	var status PermissionStatus
	var err error

	switch runtime.GOOS {
	case "darwin":
		status, err = m.requestMacOSRemoteControlPermission()
	case "windows":
		status, err = m.requestWindowsRemoteControlPermission()
	case "linux":
		status, err = m.requestLinuxRemoteControlPermission()
	default:
		return Unknown, errors.New("unsupported platform for remote control")
	}

	if err == nil {
		m.permissions[RemoteControl] = status
	}
	return status, err
}

// checkRemoteControlPermission checks remote control permission based on the platform
func (m *DefaultManager) checkRemoteControlPermission() (PermissionStatus, error) {
	var status PermissionStatus
	var err error

	switch runtime.GOOS {
	case "darwin":
		status, err = m.checkMacOSRemoteControlPermission()
	case "windows":
		status, err = m.checkWindowsRemoteControlPermission()
	case "linux":
		status, err = m.checkLinuxRemoteControlPermission()
	default:
		return Unknown, errors.New("unsupported platform for remote control")
	}

	if err == nil {
		m.permissions[RemoteControl] = status
	}
	return status, err
}

// macOS permission methods
func (m *DefaultManager) checkMacOSScreenSharePermission() (PermissionStatus, error) {
	// Check screen recording permission
	cmd := exec.Command("bash", "-c", `osascript -e 'tell application "System Events" to get every process' &>/dev/null`)
	err := cmd.Run()
	if err != nil {
		if m.verbose {
			log.Printf("Screen recording permission check failed: %v", err)
		}
		return Denied, nil
	}
	return Granted, nil
}

func (m *DefaultManager) requestMacOSScreenSharePermission() (PermissionStatus, error) {
	// First check if we already have permission
	status, _ := m.checkMacOSScreenSharePermission()
	if status == Granted {
		return Granted, nil
	}

	// Request screen recording permission
	log.Println("=================================================================")
	log.Println("🔒 SCREEN RECORDING PERMISSION REQUIRED 🔒")
	log.Println("=================================================================")
	log.Println("This application needs screen recording permission to capture screenshots.")
	log.Println("")
	log.Println("Why this is needed:")
	log.Println("- To capture screenshots of your screen")
	log.Println("- To send these screenshots through the WebSocket connection")
	log.Println("")
	log.Println("Please follow these steps:")
	log.Println("1. Go to System Preferences > Security & Privacy > Privacy > Screen Recording")
	log.Println("2. Click the lock icon to make changes (you may need to enter your password)")
	log.Println("3. Add this application to the list of allowed apps or check its checkbox if already listed")
	log.Println("4. Return to this application after granting permission")
	log.Println("=================================================================")

	// Open the System Preferences to the correct pane
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture")
	err := cmd.Run()
	if err != nil {
		return Denied, fmt.Errorf("failed to open System Preferences: %w", err)
	}

	// Ask the user if they want to continue after granting permission
	log.Println("Press Enter after granting permission to try again, or Ctrl+C to exit...")

	// Wait for user input
	var input string
	fmt.Scanln(&input)

	// Check again after user input
	status, _ = m.checkMacOSScreenSharePermission()
	if status == Granted {
		log.Println("✅ Screen recording permission granted successfully!")
		return Granted, nil
	}

	log.Println("⚠️ Screen recording permission still not granted.")
	log.Println("You may need to restart the application after granting permission.")
	return Requested, nil
}

func (m *DefaultManager) checkMacOSRemoteControlPermission() (PermissionStatus, error) {
	// Check accessibility permission (required for keyboard and mouse control)
	cmd := exec.Command("bash", "-c", `osascript -e 'tell application "System Events" to keystroke ""' &>/dev/null`)
	err := cmd.Run()
	if err != nil {
		if m.verbose {
			log.Printf("Accessibility permission check failed: %v", err)
		}

		// Try a more specific check for RobotGo
		// This checks if we can simulate a mouse click
		cmd = exec.Command("bash", "-c", `osascript -e 'tell application "System Events" to click at {0, 0}' &>/dev/null`)
		err = cmd.Run()
		if err != nil {
			if m.verbose {
				log.Printf("Mouse control permission check failed: %v", err)
			}
			return Denied, nil
		}
	}
	return Granted, nil
}

func (m *DefaultManager) requestMacOSRemoteControlPermission() (PermissionStatus, error) {
	// First check if we already have permission
	status, _ := m.checkMacOSRemoteControlPermission()
	if status == Granted {
		return Granted, nil
	}

	// Request accessibility permission
	log.Println("=================================================================")
	log.Println("🔒 ACCESSIBILITY PERMISSION REQUIRED 🔒")
	log.Println("=================================================================")
	log.Println("This application needs accessibility permission to control the mouse and keyboard.")
	log.Println("")
	log.Println("Why this is needed:")
	log.Println("- To enable remote control functionality")
	log.Println("- To simulate mouse movements and clicks")
	log.Println("- To simulate keyboard input")
	log.Println("")
	log.Println("Please follow these steps:")
	log.Println("1. Go to System Preferences > Security & Privacy > Privacy > Accessibility")
	log.Println("2. Click the lock icon to make changes (you may need to enter your password)")
	log.Println("3. Add this application to the list of allowed apps or check its checkbox if already listed")
	log.Println("4. Return to this application after granting permission")
	log.Println("=================================================================")

	// Open the System Preferences to the correct pane
	cmd := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility")
	err := cmd.Run()
	if err != nil {
		return Denied, fmt.Errorf("failed to open System Preferences: %w", err)
	}

	// Ask the user if they want to continue after granting permission
	log.Println("Press Enter after granting permission to try again, or Ctrl+C to exit...")

	// Wait for user input
	var input string
	fmt.Scanln(&input)

	// Check again after user input
	status, _ = m.checkMacOSRemoteControlPermission()
	if status == Granted {
		log.Println("✅ Accessibility permission granted successfully!")
		return Granted, nil
	}

	log.Println("⚠️ Accessibility permission still not granted.")
	log.Println("You may need to restart the application after granting permission.")
	return Requested, nil
}

// Windows permission methods
func (m *DefaultManager) checkWindowsScreenSharePermission() (PermissionStatus, error) {
	// Windows doesn't have a permission system like macOS
	// For screen capture, we'll just return Granted
	return Granted, nil
}

func (m *DefaultManager) requestWindowsScreenSharePermission() (PermissionStatus, error) {
	// Windows doesn't have a permission system like macOS
	// For screen capture, we'll just return Granted
	return Granted, nil
}

func (m *DefaultManager) checkWindowsRemoteControlPermission() (PermissionStatus, error) {
	// Windows doesn't have a permission system like macOS
	// For remote control, we'll just return Granted
	return Granted, nil
}

func (m *DefaultManager) requestWindowsRemoteControlPermission() (PermissionStatus, error) {
	// Windows doesn't have a permission system like macOS
	// For remote control, we'll just return Granted
	return Granted, nil
}

// Linux permission methods
func (m *DefaultManager) checkLinuxScreenSharePermission() (PermissionStatus, error) {
	// Check if we can access the X server
	cmd := exec.Command("xdpyinfo")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if m.verbose {
			log.Printf("xdpyinfo output: %s", string(output))
		}
		return Denied, nil
	}
	return Granted, nil
}

func (m *DefaultManager) requestLinuxScreenSharePermission() (PermissionStatus, error) {
	// First check if we already have permission
	status, _ := m.checkLinuxScreenSharePermission()
	if status == Granted {
		return Granted, nil
	}

	// For Linux, we need to ensure X11 access
	log.Println("This application requires access to the X server for screen capture.")
	log.Println("If running via SSH, make sure to enable X11 forwarding.")
	log.Println("If running locally, ensure the DISPLAY environment variable is set correctly.")

	// Check if we're running in a Wayland session
	cmd := exec.Command("bash", "-c", "echo $XDG_SESSION_TYPE")
	output, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(output)) == "wayland" {
		log.Println("Warning: You are running in a Wayland session. Screen capture may not work correctly.")
		log.Println("Consider switching to an X11 session for better compatibility.")
	}

	return Requested, nil
}

func (m *DefaultManager) checkLinuxRemoteControlPermission() (PermissionStatus, error) {
	// Check if we can access the X server for input
	cmd := exec.Command("xset", "q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if m.verbose {
			log.Printf("xset output: %s", string(output))
		}
		return Denied, nil
	}
	return Granted, nil
}

func (m *DefaultManager) requestLinuxRemoteControlPermission() (PermissionStatus, error) {
	// First check if we already have permission
	status, _ := m.checkLinuxRemoteControlPermission()
	if status == Granted {
		return Granted, nil
	}

	// For Linux, we need to ensure X11 access
	log.Println("This application requires access to the X server for remote control.")
	log.Println("If running via SSH, make sure to enable X11 forwarding.")
	log.Println("If running locally, ensure the DISPLAY environment variable is set correctly.")

	// Check if we're running in a Wayland session
	cmd := exec.Command("bash", "-c", "echo $XDG_SESSION_TYPE")
	output, err := cmd.CombinedOutput()
	if err == nil && strings.TrimSpace(string(output)) == "wayland" {
		log.Println("Warning: You are running in a Wayland session. Remote control may not work correctly.")
		log.Println("Consider switching to an X11 session for better compatibility.")
	}

	return Requested, nil
}

// RequestPermissionInteractive requests a permission with an interactive flow
// It returns true if the permission was granted, false otherwise
func (m *DefaultManager) RequestPermissionInteractive(permType PermissionType) bool {
	// First check if we already have the permission
	status, err := m.CheckPermission(permType)
	if err == nil && status == Granted {
		return true
	}

	// Start interactive flow
	fmt.Println("\n=================================================================")
	fmt.Printf("🔒 PERMISSION REQUEST: %s 🔒\n", permType)
	fmt.Println("=================================================================")

	var description, instructions string
	var preferencesPath string

	switch permType {
	case ScreenShare:
		description = "Screen recording permission is required to capture screenshots."
		instructions = "Please follow these steps:\n" +
			"1. Go to System Preferences > Security & Privacy > Privacy > Screen Recording\n" +
			"2. Click the lock icon to make changes (you may need to enter your password)\n" +
			"3. Add this application to the list of allowed apps or check its checkbox if already listed\n" +
			"4. Return to this application after granting permission"
		preferencesPath = "x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture"
	case RemoteControl:
		description = "Accessibility permission is required to control the mouse and keyboard."
		instructions = "Please follow these steps:\n" +
			"1. Go to System Preferences > Security & Privacy > Privacy > Accessibility\n" +
			"2. Click the lock icon to make changes (you may need to enter your password)\n" +
			"3. Add this application to the list of allowed apps or check its checkbox if already listed\n" +
			"4. Return to this application after granting permission"
		preferencesPath = "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
	default:
		fmt.Printf("Unknown permission type: %s\n", permType)
		return false
	}

	// Print description and instructions
	fmt.Println("\n" + description)
	fmt.Println("\n" + instructions)

	// Ask user if they want to open System Preferences
	fmt.Println("\nWould you like to open System Preferences now? (y/n)")
	var input string
	fmt.Scanln(&input)

	if input == "y" || input == "Y" {
		// Open System Preferences
		if runtime.GOOS == "darwin" {
			cmd := exec.Command("open", preferencesPath)
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Failed to open System Preferences: %v\n", err)
			} else {
				fmt.Println("System Preferences opened. Please grant the permission.")
			}
		} else if runtime.GOOS == "linux" {
			fmt.Println("On Linux, you may need to run this application with sudo or adjust permissions manually.")
		} else if runtime.GOOS == "windows" {
			fmt.Println("On Windows, you typically don't need special permissions for these operations.")
		}
	}

	// Wait for user to grant permission
	fmt.Println("\nPress Enter after granting permission to check again, or type 'skip' to continue without permission.")
	fmt.Scanln(&input)

	if input == "skip" {
		fmt.Println("Continuing without permission. Some features may not work correctly.")
		return false
	}

	// Check if permission was granted
	status, _ = m.CheckPermission(permType)
	if status == Granted {
		fmt.Println("\n✅ Permission granted successfully!")
		return true
	} else {
		fmt.Println("\n❌ Permission was not granted.")
		fmt.Println("Would you like to try again? (y/n)")
		fmt.Scanln(&input)

		if input == "y" || input == "Y" {
			// Recursive call to try again
			return m.RequestPermissionInteractive(permType)
		} else {
			fmt.Println("Continuing without permission. Some features may not work correctly.")
			return false
		}
	}
}
