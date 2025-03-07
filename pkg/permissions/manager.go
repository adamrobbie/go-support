package permissions

import (
	"errors"
	"fmt"
	"runtime"
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
}

// DefaultManager is the default implementation of Manager
type DefaultManager struct {
	permissions map[PermissionType]PermissionStatus
}

// NewManager creates a new permission manager
func NewManager() Manager {
	return &DefaultManager{
		permissions: make(map[PermissionType]PermissionStatus),
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
	default:
		return Unknown, fmt.Errorf("unsupported permission type: %s", permType)
	}
}

// CheckPermission implements the Manager interface
func (m *DefaultManager) CheckPermission(permType PermissionType) (PermissionStatus, error) {
	status, exists := m.permissions[permType]
	if !exists {
		return Unknown, nil
	}
	return status, nil
}

// requestScreenSharePermission requests screen sharing permission based on the platform
func (m *DefaultManager) requestScreenSharePermission() (PermissionStatus, error) {
	var status PermissionStatus
	var err error

	switch runtime.GOOS {
	case "darwin":
		status, err = requestMacOSScreenSharePermission()
	case "windows":
		status, err = requestWindowsScreenSharePermission()
	case "linux":
		status, err = requestLinuxScreenSharePermission()
	default:
		return Unknown, errors.New("unsupported platform for screen sharing")
	}

	if err == nil {
		m.permissions[ScreenShare] = status
	}
	return status, err
}
