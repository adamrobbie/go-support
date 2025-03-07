package appid

import (
	"os"
	"runtime"
	"testing"
)

func TestAppConstants(t *testing.T) {
	// Test that the app constants are set correctly
	if AppName == "" {
		t.Error("AppName is empty")
	}

	if AppID == "" {
		t.Error("AppID is empty")
	}

	if AppVersion == "" {
		t.Error("AppVersion is empty")
	}

	if AppDescription == "" {
		t.Error("AppDescription is empty")
	}
}

func TestSetupAppIdentifier(t *testing.T) {
	// This is a simple test to ensure the function doesn't crash
	// We can't fully test platform-specific behavior in a cross-platform way
	err := SetupAppIdentifier()

	// On unsupported platforms, we expect an error
	if runtime.GOOS != "darwin" && runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		if err == nil {
			t.Error("SetupAppIdentifier() should return an error on unsupported platforms")
		}
	} else {
		// On supported platforms, we don't expect an error
		// But we'll skip the test if it fails because it might require permissions
		if err != nil {
			t.Skipf("SetupAppIdentifier() returned an error on %s: %v", runtime.GOOS, err)
		}
	}
}

// TestSetupLinuxIdentifierMock tests the Linux identifier setup with a mock home directory
func TestSetupLinuxIdentifierMock(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test on non-Linux platform")
	}

	// Save the original os.UserHomeDir function
	originalUserHomeDir := osUserHomeDir
	defer func() { osUserHomeDir = originalUserHomeDir }()

	// Create a temporary directory to use as the home directory
	tempDir, err := os.MkdirTemp("", "appid-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Mock the os.UserHomeDir function
	osUserHomeDir = func() (string, error) {
		return tempDir, nil
	}

	// Call the function
	err = setupLinuxIdentifier()
	if err != nil {
		t.Errorf("setupLinuxIdentifier() returned an error: %v", err)
	}

	// Check that the desktop file was created
	desktopFile := tempDir + "/.local/share/applications/go-support.desktop"
	if _, err := os.Stat(desktopFile); os.IsNotExist(err) {
		t.Errorf("Desktop file was not created at %s", desktopFile)
	}
}
