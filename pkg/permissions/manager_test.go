package permissions

import (
	"errors"
	"os/exec"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(false)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Check that the manager is of the correct type
	defaultManager, ok := manager.(*DefaultManager)
	if !ok {
		t.Errorf("NewManager() returned wrong type: %T", manager)
	}

	// Check that the verbose flag is set correctly
	if defaultManager.verbose {
		t.Error("Expected verbose to be false")
	}

	// Test with verbose=true
	verboseManager := NewManager(true)
	verboseDefaultManager, ok := verboseManager.(*DefaultManager)
	if !ok {
		t.Errorf("NewManager(true) returned wrong type: %T", verboseManager)
	}

	if !verboseDefaultManager.verbose {
		t.Error("Expected verbose to be true")
	}
}

func TestDefaultManagerCheckPermission(t *testing.T) {
	// Save original exec functions
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	// Mock the exec.Command function to simulate permission check failure
	execCommand = func(command string, args ...string) *exec.Cmd {
		// Return a command that will fail, simulating no permission
		return exec.Command("echo", "permission denied")
	}

	manager := &DefaultManager{
		permissions: make(map[PermissionType]PermissionStatus),
		verbose:     false,
	}

	// Test checking a permission that doesn't exist in the cache
	// This will call the platform-specific function which we've mocked
	status, err := manager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}

	// Now manually set permissions in the cache for testing

	// Set a permission in the cache and test checking it
	manager.permissions[ScreenShare] = Granted
	status, err = manager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}
	if status != Granted {
		t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, Granted)
	}

	// Test with different permission statuses
	testCases := []struct {
		name   string
		status PermissionStatus
	}{
		{"Denied", Denied},
		{"Requested", Requested},
		{"Unknown", Unknown},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager.permissions[ScreenShare] = tc.status
			status, err := manager.CheckPermission(ScreenShare)
			if err != nil {
				t.Errorf("CheckPermission() returned an error: %v", err)
			}
			if status != tc.status {
				t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, tc.status)
			}
		})
	}
}

func TestDefaultManagerRequestPermission(t *testing.T) {
	// Create a manager with a permission already granted
	manager := &DefaultManager{
		permissions: map[PermissionType]PermissionStatus{
			ScreenShare: Granted,
		},
	}

	// Test requesting a permission that's already granted
	status, err := manager.RequestPermission(ScreenShare)
	if err != nil {
		t.Errorf("RequestPermission() returned an error: %v", err)
	}
	if status != Granted {
		t.Errorf("RequestPermission() returned wrong status: got %v, want %v", status, Granted)
	}

	// Test requesting an unsupported permission type
	unsupportedType := PermissionType("unsupported")
	status, err = manager.RequestPermission(unsupportedType)
	if err == nil {
		t.Error("RequestPermission() with unsupported type should return an error")
	}
	if status != Unknown {
		t.Errorf("RequestPermission() with unsupported type returned wrong status: got %v, want %v", status, Unknown)
	}
}

func TestMockManager(t *testing.T) {
	mockManager := NewMockManager()

	// Test default behavior
	status, err := mockManager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}
	if status != Unknown {
		t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, Unknown)
	}

	// Set a permission and test checking it
	mockManager.SetPermission(ScreenShare, Granted)
	status, err = mockManager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}
	if status != Granted {
		t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, Granted)
	}

	// Test custom check function
	mockManager.SetCheckFunc(func(permType PermissionType) (PermissionStatus, error) {
		if permType == ScreenShare {
			return Denied, nil
		}
		return Unknown, errors.New("unsupported permission type")
	})

	status, err = mockManager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() with custom function returned an error: %v", err)
	}
	if status != Denied {
		t.Errorf("CheckPermission() with custom function returned wrong status: got %v, want %v", status, Denied)
	}

	// Test custom request function
	mockManager.SetRequestFunc(func(permType PermissionType) (PermissionStatus, error) {
		if permType == ScreenShare {
			return Requested, nil
		}
		return Unknown, errors.New("unsupported permission type")
	})

	status, err = mockManager.RequestPermission(ScreenShare)
	if err != nil {
		t.Errorf("RequestPermission() with custom function returned an error: %v", err)
	}
	if status != Requested {
		t.Errorf("RequestPermission() with custom function returned wrong status: got %v, want %v", status, Requested)
	}

	// Test error cases
	mockManager.SetCheckFunc(func(permType PermissionType) (PermissionStatus, error) {
		return Unknown, errors.New("check error")
	})

	_, err = mockManager.CheckPermission(ScreenShare)
	if err == nil {
		t.Error("CheckPermission() with error function should return an error")
	}

	mockManager.SetRequestFunc(func(permType PermissionType) (PermissionStatus, error) {
		return Unknown, errors.New("request error")
	})

	_, err = mockManager.RequestPermission(ScreenShare)
	if err == nil {
		t.Error("RequestPermission() with error function should return an error")
	}
}

// TestPermissionTypeString tests the string representation of PermissionType
func TestPermissionTypeString(t *testing.T) {
	if ScreenShare.String() != "screen_share" {
		t.Errorf("ScreenShare.String() = %v, want %v", ScreenShare.String(), "screen_share")
	}

	customType := PermissionType("custom_type")
	if customType.String() != "custom_type" {
		t.Errorf("customType.String() = %v, want %v", customType.String(), "custom_type")
	}
}

// TestPermissionStatusString tests the string representation of PermissionStatus
func TestPermissionStatusString(t *testing.T) {
	testCases := []struct {
		status PermissionStatus
		want   string
	}{
		{Unknown, "Unknown"},
		{Granted, "Granted"},
		{Denied, "Denied"},
		{Requested, "Requested"},
		{PermissionStatus(99), "Unknown Status: 99"},
	}

	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			if tc.status.String() != tc.want {
				t.Errorf("PermissionStatus(%d).String() = %v, want %v", tc.status, tc.status.String(), tc.want)
			}
		})
	}
}

// TestEnsurePermission tests the EnsurePermission method
func TestEnsurePermission(t *testing.T) {
	// Create a mock manager for testing
	mockManager := NewMockManager()

	// Test with a permission that's already granted
	mockManager.SetPermission(ScreenShare, Granted)
	granted, err := mockManager.EnsurePermission(ScreenShare)
	if err != nil {
		t.Errorf("EnsurePermission() returned an error: %v", err)
	}
	if !granted {
		t.Error("EnsurePermission() should return true for a granted permission")
	}

	// Test with a permission that's denied
	mockManager.SetPermission(ScreenShare, Denied)
	granted, err = mockManager.EnsurePermission(ScreenShare)
	if err != nil {
		t.Errorf("EnsurePermission() returned an error: %v", err)
	}
	if granted {
		t.Error("EnsurePermission() should return false for a denied permission")
	}

	// Test with a permission that's not in the cache
	delete(mockManager.permissions, ScreenShare)

	// Set up the mock to return Denied for RequestPermission
	mockManager.SetRequestFunc(func(permType PermissionType) (PermissionStatus, error) {
		return Denied, nil
	})

	granted, err = mockManager.EnsurePermission(ScreenShare)
	if err != nil {
		t.Errorf("EnsurePermission() returned an error: %v", err)
	}
	if granted {
		t.Error("EnsurePermission() should return false when RequestPermission returns Denied")
	}

	// Test with an error from RequestPermission
	mockManager.SetRequestFunc(func(permType PermissionType) (PermissionStatus, error) {
		return Unknown, errors.New("request error")
	})

	_, err = mockManager.EnsurePermission(ScreenShare)
	if err == nil {
		t.Error("EnsurePermission() should return an error when RequestPermission fails")
	}
}

// TestRequestPermissionInteractive tests the RequestPermissionInteractive method
func TestRequestPermissionInteractive(t *testing.T) {
	// Create a mock manager for testing
	mockManager := NewMockManager()

	// Test with a permission that's already granted
	mockManager.SetPermission(ScreenShare, Granted)
	granted := mockManager.RequestPermissionInteractive(ScreenShare)
	if !granted {
		t.Error("RequestPermissionInteractive() should return true for a granted permission")
	}

	// Test with a permission that's denied
	mockManager.SetPermission(ScreenShare, Denied)
	granted = mockManager.RequestPermissionInteractive(ScreenShare)
	if granted {
		t.Error("RequestPermissionInteractive() should return false for a denied permission")
	}
}
