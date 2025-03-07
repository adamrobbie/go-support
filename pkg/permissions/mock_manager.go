package permissions

// MockManager is a mock implementation of the Manager interface for testing
type MockManager struct {
	permissions map[PermissionType]PermissionStatus
	requestFunc func(PermissionType) (PermissionStatus, error)
	checkFunc   func(PermissionType) (PermissionStatus, error)
}

// NewMockManager creates a new mock permission manager
func NewMockManager() *MockManager {
	return &MockManager{
		permissions: make(map[PermissionType]PermissionStatus),
		requestFunc: nil,
		checkFunc:   nil,
	}
}

// SetRequestFunc sets the function to be called when RequestPermission is called
func (m *MockManager) SetRequestFunc(f func(PermissionType) (PermissionStatus, error)) {
	m.requestFunc = f
}

// SetCheckFunc sets the function to be called when CheckPermission is called
func (m *MockManager) SetCheckFunc(f func(PermissionType) (PermissionStatus, error)) {
	m.checkFunc = f
}

// SetPermission sets the permission status for a specific permission type
func (m *MockManager) SetPermission(permType PermissionType, status PermissionStatus) {
	m.permissions[permType] = status
}

// RequestPermission implements the Manager interface
func (m *MockManager) RequestPermission(permType PermissionType) (PermissionStatus, error) {
	if m.requestFunc != nil {
		return m.requestFunc(permType)
	}

	// Default implementation
	status, exists := m.permissions[permType]
	if !exists {
		return Unknown, nil
	}
	return status, nil
}

// CheckPermission implements the Manager interface
func (m *MockManager) CheckPermission(permType PermissionType) (PermissionStatus, error) {
	if m.checkFunc != nil {
		return m.checkFunc(permType)
	}

	// Default implementation
	status, exists := m.permissions[permType]
	if !exists {
		return Unknown, nil
	}
	return status, nil
}
