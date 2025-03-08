//go:build test
// +build test

package remote

import (
	"fmt"
	"sync"
)

// Mock variables for RobotGo functions
var (
	// Mutex to protect the mock variables
	mockMutex sync.Mutex

	// Mock return values
	mockScreenWidth  = 1920
	mockScreenHeight = 1080
	mockMouseX       = 500
	mockMouseY       = 500

	// Mock errors
	mockMoveMouseErr     error
	mockClickErr         error
	mockToggleErr        error
	mockTypeStringErr    error
	mockKeyTapErr        error
	mockGetScreenSizeErr error
	mockGetMousePosErr   error

	// Call tracking
	moveMouseCalled     bool
	clickCalled         bool
	toggleCalled        bool
	typeStringCalled    bool
	keyTapCalled        bool
	getScreenSizeCalled bool
	getMousePosCalled   bool

	// Call arguments
	lastMoveMouseX      int
	lastMoveMouseY      int
	lastClickButton     string
	lastClickDouble     bool
	lastToggleButton    string
	lastToggleDirection string
	lastTypeString      string
	lastKeyTap          string

	// Mock implementations
	robotgoMoveMouseFunc = func(x, y int) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		moveMouseCalled = true
		lastMoveMouseX = x
		lastMoveMouseY = y
		mockMouseX = x
		mockMouseY = y
	}

	robotgoClickFunc = func(button string, double bool) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		clickCalled = true
		lastClickButton = button
		lastClickDouble = double
	}

	robotgoToggleFunc = func(button, direction string) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		toggleCalled = true
		lastToggleButton = button
		lastToggleDirection = direction
	}

	robotgoTypeStringFunc = func(text string) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		typeStringCalled = true
		lastTypeString = text
	}

	robotgoKeyTapFunc = func(key string, modifiers ...string) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		keyTapCalled = true
		lastKeyTap = key
	}

	robotgoGetScreenSizeFunc = func() (int, int) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		getScreenSizeCalled = true
		return mockScreenWidth, mockScreenHeight
	}

	robotgoGetMousePosFunc = func() (int, int) {
		mockMutex.Lock()
		defer mockMutex.Unlock()
		getMousePosCalled = true
		return mockMouseX, mockMouseY
	}
)

// ResetMocks resets all mock variables
func ResetMocks() {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	// Reset return values
	mockScreenWidth = 1920
	mockScreenHeight = 1080
	mockMouseX = 500
	mockMouseY = 500

	// Reset errors
	mockMoveMouseErr = nil
	mockClickErr = nil
	mockToggleErr = nil
	mockTypeStringErr = nil
	mockKeyTapErr = nil
	mockGetScreenSizeErr = nil
	mockGetMousePosErr = nil

	// Reset call tracking
	moveMouseCalled = false
	clickCalled = false
	toggleCalled = false
	typeStringCalled = false
	keyTapCalled = false
	getScreenSizeCalled = false
	getMousePosCalled = false

	// Reset call arguments
	lastMoveMouseX = 0
	lastMoveMouseY = 0
	lastClickButton = ""
	lastClickDouble = false
	lastToggleButton = ""
	lastToggleDirection = ""
	lastTypeString = ""
	lastKeyTap = ""
}

// SetMockScreenSize sets the mock screen size
func SetMockScreenSize(width, height int) {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	mockScreenWidth = width
	mockScreenHeight = height
}

// SetMockMousePosition sets the mock mouse position
func SetMockMousePosition(x, y int) {
	mockMutex.Lock()
	defer mockMutex.Unlock()
	mockMouseX = x
	mockMouseY = y
}

// SetMockError sets a mock error for a specific function
func SetMockError(function string, err error) {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	switch function {
	case "MoveMouse":
		mockMoveMouseErr = err
	case "Click":
		mockClickErr = err
	case "Toggle":
		mockToggleErr = err
	case "TypeString":
		mockTypeStringErr = err
	case "KeyTap":
		mockKeyTapErr = err
	case "GetScreenSize":
		mockGetScreenSizeErr = err
	case "GetMousePos":
		mockGetMousePosErr = err
	default:
		panic(fmt.Sprintf("Unknown function: %s", function))
	}
}

// GetMockCallCount returns whether a specific function was called
func GetMockCallCount(function string) bool {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	switch function {
	case "MoveMouse":
		return moveMouseCalled
	case "Click":
		return clickCalled
	case "Toggle":
		return toggleCalled
	case "TypeString":
		return typeStringCalled
	case "KeyTap":
		return keyTapCalled
	case "GetScreenSize":
		return getScreenSizeCalled
	case "GetMousePos":
		return getMousePosCalled
	default:
		panic(fmt.Sprintf("Unknown function: %s", function))
	}
}

// GetMockLastArgs returns the last arguments passed to a specific function
func GetMockLastArgs(function string) map[string]interface{} {
	mockMutex.Lock()
	defer mockMutex.Unlock()

	switch function {
	case "MoveMouse":
		return map[string]interface{}{
			"x": lastMoveMouseX,
			"y": lastMoveMouseY,
		}
	case "Click":
		return map[string]interface{}{
			"button": lastClickButton,
			"double": lastClickDouble,
		}
	case "Toggle":
		return map[string]interface{}{
			"button":    lastToggleButton,
			"direction": lastToggleDirection,
		}
	case "TypeString":
		return map[string]interface{}{
			"text": lastTypeString,
		}
	case "KeyTap":
		return map[string]interface{}{
			"key": lastKeyTap,
		}
	default:
		panic(fmt.Sprintf("Unknown function: %s", function))
	}
}
