package appid

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CreateMacOSInfoPlist creates an Info.plist file for macOS
func CreateMacOSInfoPlist(executablePath string) error {
	// Get the directory of the executable
	execDir := filepath.Dir(executablePath)

	// Create Contents directory if it doesn't exist
	contentsDir := filepath.Join(execDir, "Contents")
	if err := os.MkdirAll(contentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create Contents directory: %w", err)
	}

	// Create MacOS directory if it doesn't exist
	macosDir := filepath.Join(contentsDir, "MacOS")
	if err := os.MkdirAll(macosDir, 0755); err != nil {
		return fmt.Errorf("failed to create MacOS directory: %w", err)
	}

	// Create Resources directory if it doesn't exist
	resourcesDir := filepath.Join(contentsDir, "Resources")
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create Resources directory: %w", err)
	}

	// Create Info.plist file
	infoPlistPath := filepath.Join(contentsDir, "Info.plist")
	infoPlistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>%s</string>
	<key>CFBundleName</key>
	<string>%s</string>
	<key>CFBundleDisplayName</key>
	<string>%s</string>
	<key>CFBundleVersion</key>
	<string>%s</string>
	<key>CFBundleShortVersionString</key>
	<string>%s</string>
	<key>CFBundleExecutable</key>
	<string>%s</string>
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>NSSupportsAutomaticGraphicsSwitching</key>
	<true/>
</dict>
</plist>`, AppID, AppName, AppName, AppVersion, AppVersion, filepath.Base(executablePath))

	if err := os.WriteFile(infoPlistPath, []byte(infoPlistContent), 0644); err != nil {
		return fmt.Errorf("failed to write Info.plist file: %w", err)
	}

	return nil
}

// RegisterMacOSAppWithLaunchServices registers the application with Launch Services
func RegisterMacOSAppWithLaunchServices(appPath string) error {
	// Use the lsregister command to register the application
	cmd := exec.Command("/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister", "-f", appPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to register application with Launch Services: %w, output: %s", err, string(output))
	}

	return nil
}

// GetMacOSBundleIdentifier gets the bundle identifier for the current process
func GetMacOSBundleIdentifier() (string, error) {
	// Use the defaults command to get the bundle identifier
	cmd := exec.Command("defaults", "read", filepath.Join(os.Args[0], "Contents", "Info"), "CFBundleIdentifier")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the command fails, return the default bundle identifier
		return AppID, nil
	}

	return strings.TrimSpace(string(output)), nil
}

// SetupMacOSAppBundle sets up a macOS application bundle
func SetupMacOSAppBundle() error {
	// Get the absolute path of the executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Check if we're already in an app bundle
	if strings.Contains(execPath, ".app/Contents/MacOS") {
		fmt.Println("Already running from an application bundle")
		return nil
	}

	// Create app bundle directory
	appBundlePath := execPath + ".app"
	fmt.Printf("Creating application bundle at %s\n", appBundlePath)

	// Create Info.plist
	if err := CreateMacOSInfoPlist(execPath); err != nil {
		return fmt.Errorf("failed to create Info.plist: %w", err)
	}

	fmt.Printf("Application bundle created with identifier: %s\n", AppID)

	return nil
}
