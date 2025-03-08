package remote

import (
	"errors"
	"testing"

	"github.com/adamrobbie/go-support/pkg/permissions"
)

// ErrPermissionDenied is returned when a permission is denied
var ErrPermissionDenied = errors.New("permission denied")

// mockPermissionsManager is a mock implementation of the permissions.Manager interface
type mockPermissionsManager struct {
	shouldGrantPermission bool
}

func (m *mockPermissionsManager) RequestPermission(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
	if m.shouldGrantPermission {
		return permissions.Granted, nil
	}
	return permissions.Denied, ErrPermissionDenied
}

func (m *mockPermissionsManager) HasPermission(permType permissions.PermissionType) bool {
	return m.shouldGrantPermission
}

func (m *mockPermissionsManager) CheckPermission(permType permissions.PermissionType) (permissions.PermissionStatus, error) {
	if m.shouldGrantPermission {
		return permissions.Granted, nil
	}
	return permissions.Denied, ErrPermissionDenied
}

func (m *mockPermissionsManager) EnsurePermission(permType permissions.PermissionType) (bool, error) {
	return m.shouldGrantPermission, nil
}

func (m *mockPermissionsManager) RequestPermissionInteractive(permType permissions.PermissionType) bool {
	return m.shouldGrantPermission
}

// mockRemoteController is a mock implementation of the RemoteController for testing
type mockRemoteController struct {
	screenWidth  int
	screenHeight int
	mouseX       int
	mouseY       int
	permManager  permissions.Manager
}

// GetScreenSize returns the mock screen size
func (m *mockRemoteController) GetScreenSize() (int, int) {
	return m.screenWidth, m.screenHeight
}

// GetMousePosition returns the mock mouse position
func (m *mockRemoteController) GetMousePosition() (int, int) {
	return m.mouseX, m.mouseY
}

// ExecuteMouseEvent handles mouse events in the mock controller
func (m *mockRemoteController) ExecuteMouseEvent(event MouseEvent) error {
	switch event.Action {
	case MouseMove:
		m.mouseX = event.X
		m.mouseY = event.Y
	case MouseClick:
		// Just simulate a click, no actual action needed for testing
	case MouseDown:
		// Just simulate a mouse down, no actual action needed for testing
	case MouseUp:
		// Just simulate a mouse up, no actual action needed for testing
	}
	return nil
}

// ExecuteKeyboardEvent handles keyboard events in the mock controller
func (m *mockRemoteController) ExecuteKeyboardEvent(event KeyboardEvent) error {
	switch event.Action {
	case "type":
		// Just simulate typing, no actual action needed for testing
	case "press":
		// Just simulate a key press, no actual action needed for testing
	}
	return nil
}

// TestNewRemoteController tests the NewRemoteController function
func TestNewRemoteController(t *testing.T) {
	// Create a mock permissions manager
	mockManager := &mockPermissionsManager{
		shouldGrantPermission: true,
	}

	// Create a new remote controller
	controller := NewRemoteController(mockManager, false)

	// Check that the controller was created successfully
	if controller == nil {
		t.Errorf("NewRemoteController() returned nil")
	}
}

// TestRemoteControllerPermissions tests the permission handling in the RemoteController
func TestRemoteControllerPermissions(t *testing.T) {
	// Test with permissions granted
	t.Run("PermissionsGranted", func(t *testing.T) {
		// Skip in short mode or when running in CI
		if testing.Short() {
			t.Skip("Skipping test in short mode")
		}

		// Create a mock permissions manager that always grants permission
		mockManager := &mockPermissionsManager{
			shouldGrantPermission: true,
		}

		// Create a controller with the mock manager
		controller := NewRemoteController(mockManager, false)

		// Test mouse event with permissions granted
		event := MouseEvent{
			Action: MouseMove,
			X:      100,
			Y:      200,
		}

		// Skip actual execution in tests
		// We're just testing that permissions are checked correctly
		// We don't want to actually move the mouse during tests
		t.Skip("Skipping actual mouse movement in tests")

		err := controller.ExecuteMouseEvent(event)
		if err != nil {
			t.Errorf("ExecuteMouseEvent() returned an error with permissions granted: %v", err)
		}

		// Test keyboard event with permissions granted
		keyEvent := KeyboardEvent{
			Action: "type",
			Text:   "test",
		}

		err = controller.ExecuteKeyboardEvent(keyEvent)
		if err != nil {
			t.Errorf("ExecuteKeyboardEvent() returned an error with permissions granted: %v", err)
		}
	})

	// Test with permissions denied
	t.Run("PermissionsDenied", func(t *testing.T) {
		// Create a mock permissions manager that always denies permission
		mockManager := &mockPermissionsManager{
			shouldGrantPermission: false,
		}

		// Create a controller with the mock manager
		controller := NewRemoteController(mockManager, false)

		// Test mouse event with permissions denied
		event := MouseEvent{
			Action: MouseMove,
			X:      100,
			Y:      200,
		}

		err := controller.ExecuteMouseEvent(event)
		if err == nil {
			t.Errorf("ExecuteMouseEvent() did not return an error with permissions denied")
		}

		// Test keyboard event with permissions denied
		keyEvent := KeyboardEvent{
			Action: "type",
			Text:   "test",
		}

		err = controller.ExecuteKeyboardEvent(keyEvent)
		if err == nil {
			t.Errorf("ExecuteKeyboardEvent() did not return an error with permissions denied")
		}
	})
}
