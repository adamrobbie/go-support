package permissions

import (
	"errors"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Check that the manager is of the correct type
	_, ok := manager.(*DefaultManager)
	if !ok {
		t.Errorf("NewManager() returned wrong type: %T", manager)
	}
}

func TestDefaultManagerCheckPermission(t *testing.T) {
	manager := &DefaultManager{
		permissions: make(map[PermissionType]PermissionStatus),
	}

	// Test checking a permission that doesn't exist
	status, err := manager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}
	if status != Unknown {
		t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, Unknown)
	}

	// Set a permission and test checking it
	manager.permissions[ScreenShare] = Granted
	status, err = manager.CheckPermission(ScreenShare)
	if err != nil {
		t.Errorf("CheckPermission() returned an error: %v", err)
	}
	if status != Granted {
		t.Errorf("CheckPermission() returned wrong status: got %v, want %v", status, Granted)
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
}
