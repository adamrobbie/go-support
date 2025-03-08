//go:build test
// +build test

package remote

import (
	"errors"
	"testing"
)

// TestGetScreenSizeHelper tests the GetScreenSize helper function
func TestGetScreenSizeHelper(t *testing.T) {
	// Save original function and restore it after the test
	originalFunc := robotgoGetScreenSizeFunc
	defer func() { robotgoGetScreenSizeFunc = originalFunc }()

	// Reset mocks
	ResetMocks()

	// Replace the function with our mock
	robotgoGetScreenSizeFunc = func() (int, int) {
		return mockScreenWidth, mockScreenHeight
	}

	// Test with default values
	width, height := GetScreenSize()
	if width != mockScreenWidth || height != mockScreenHeight {
		t.Errorf("GetScreenSize() = (%d, %d), want (%d, %d)", width, height, mockScreenWidth, mockScreenHeight)
	}

	// Test with custom values
	SetMockScreenSize(800, 600)
	width, height = GetScreenSize()
	if width != 800 || height != 600 {
		t.Errorf("GetScreenSize() = (%d, %d), want (800, 600)", width, height)
	}

	// Verify that the function was called
	if !GetMockCallCount("GetScreenSize") {
		t.Error("GetScreenSize() did not call robotgo.GetScreenSize()")
	}
}

// TestGetMousePositionHelper tests the GetMousePosition helper function
func TestGetMousePositionHelper(t *testing.T) {
	// Save original function and restore it after the test
	originalFunc := robotgoGetMousePosFunc
	defer func() { robotgoGetMousePosFunc = originalFunc }()

	// Reset mocks
	ResetMocks()

	// Replace the function with our mock
	robotgoGetMousePosFunc = func() (int, int) {
		return mockMouseX, mockMouseY
	}

	// Test with default values
	x, y := GetMousePosition()
	if x != mockMouseX || y != mockMouseY {
		t.Errorf("GetMousePosition() = (%d, %d), want (%d, %d)", x, y, mockMouseX, mockMouseY)
	}

	// Test with custom values
	SetMockMousePosition(300, 400)
	x, y = GetMousePosition()
	if x != 300 || y != 400 {
		t.Errorf("GetMousePosition() = (%d, %d), want (300, 400)", x, y)
	}

	// Verify that the function was called
	if !GetMockCallCount("GetMousePos") {
		t.Error("GetMousePosition() did not call robotgo.GetMousePos()")
	}
}

// TestExecuteMouseEventHelper tests the ExecuteMouseEvent helper function
func TestExecuteMouseEventHelper(t *testing.T) {
	// Save original functions and restore them after the test
	originalMoveMouseFunc := robotgoMoveMouseFunc
	originalClickFunc := robotgoClickFunc
	originalToggleFunc := robotgoToggleFunc
	defer func() {
		robotgoMoveMouseFunc = originalMoveMouseFunc
		robotgoClickFunc = originalClickFunc
		robotgoToggleFunc = originalToggleFunc
	}()

	// Reset mocks
	ResetMocks()

	// Replace the functions with our mocks
	robotgoMoveMouseFunc = func(x, y int) {
		moveMouseCalled = true
		lastMoveMouseX = x
		lastMoveMouseY = y
	}

	robotgoClickFunc = func(button string, double bool) {
		clickCalled = true
		lastClickButton = button
		lastClickDouble = double
	}

	robotgoToggleFunc = func(button, direction string) {
		toggleCalled = true
		lastToggleButton = button
		lastToggleDirection = direction
	}

	// Test cases
	testCases := []struct {
		name          string
		event         MouseEvent
		expectedFunc  string
		expectedError bool
	}{
		{
			name: "Move",
			event: MouseEvent{
				Action: MouseMove,
				X:      100,
				Y:      200,
			},
			expectedFunc: "MoveMouse",
		},
		{
			name: "Click",
			event: MouseEvent{
				Action: MouseClick,
				Button: LeftButton,
			},
			expectedFunc: "Click",
		},
		{
			name: "Double Click",
			event: MouseEvent{
				Action: MouseDblClick,
				Button: RightButton,
			},
			expectedFunc: "Click",
		},
		{
			name: "Scroll",
			event: MouseEvent{
				Action: MouseScroll,
				Amount: 5,
			},
			expectedFunc: "Scroll",
		},
		{
			name: "Down",
			event: MouseEvent{
				Action: MouseDown,
				Button: MiddleButton,
			},
			expectedFunc: "Toggle",
		},
		{
			name: "Up",
			event: MouseEvent{
				Action: MouseUp,
				Button: LeftButton,
			},
			expectedFunc: "Toggle",
		},
		{
			name: "Invalid Action",
			event: MouseEvent{
				Action: MouseAction("invalid"),
			},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks for each test case
			ResetMocks()

			// Execute the mouse event
			err := ExecuteMouseEvent(tc.event)

			// Check for expected error
			if tc.expectedError && err == nil {
				t.Error("ExecuteMouseEvent() did not return an error for invalid action")
				return
			}

			if !tc.expectedError && err != nil {
				t.Errorf("ExecuteMouseEvent() returned an error: %v", err)
				return
			}

			// Skip further checks for error cases
			if tc.expectedError {
				return
			}

			// Check that the expected function was called
			switch tc.expectedFunc {
			case "MoveMouse":
				if !GetMockCallCount("MoveMouse") {
					t.Error("ExecuteMouseEvent() did not call robotgo.MoveMouse()")
				}
				args := GetMockLastArgs("MoveMouse")
				if args["x"] != tc.event.X || args["y"] != tc.event.Y {
					t.Errorf("ExecuteMouseEvent() called MoveMouse with wrong args: got (%d, %d), want (%d, %d)",
						args["x"], args["y"], tc.event.X, tc.event.Y)
				}
			case "Click":
				if !GetMockCallCount("Click") {
					t.Error("ExecuteMouseEvent() did not call robotgo.Click()")
				}
				args := GetMockLastArgs("Click")
				if args["button"] != string(tc.event.Button) {
					t.Errorf("ExecuteMouseEvent() called Click with wrong button: got %s, want %s",
						args["button"], tc.event.Button)
				}
				if tc.event.Action == MouseDblClick && !args["double"].(bool) {
					t.Error("ExecuteMouseEvent() called Click with double=false for double click")
				}
			case "Toggle":
				if !GetMockCallCount("Toggle") {
					t.Error("ExecuteMouseEvent() did not call robotgo.Toggle()")
				}
				args := GetMockLastArgs("Toggle")
				if args["button"] != string(tc.event.Button) {
					t.Errorf("ExecuteMouseEvent() called Toggle with wrong button: got %s, want %s",
						args["button"], tc.event.Button)
				}
				expectedDirection := "down"
				if tc.event.Action == MouseUp {
					expectedDirection = "up"
				}
				if args["direction"] != expectedDirection {
					t.Errorf("ExecuteMouseEvent() called Toggle with wrong direction: got %s, want %s",
						args["direction"], expectedDirection)
				}
			}
		})
	}
}

// TestExecuteKeyboardEventHelper tests the ExecuteKeyboardEvent helper function
func TestExecuteKeyboardEventHelper(t *testing.T) {
	// Save original functions and restore them after the test
	originalTypeStringFunc := robotgoTypeStringFunc
	originalKeyTapFunc := robotgoKeyTapFunc
	defer func() {
		robotgoTypeStringFunc = originalTypeStringFunc
		robotgoKeyTapFunc = originalKeyTapFunc
	}()

	// Reset mocks
	ResetMocks()

	// Replace the functions with our mocks
	robotgoTypeStringFunc = func(text string) {
		typeStringCalled = true
		lastTypeString = text
	}

	robotgoKeyTapFunc = func(key string, modifiers ...string) {
		keyTapCalled = true
		lastKeyTap = key
	}

	// Test cases
	testCases := []struct {
		name          string
		event         KeyboardEvent
		expectedFunc  string
		expectedError bool
	}{
		{
			name: "Press",
			event: KeyboardEvent{
				Action: KeyPress,
				Key:    "escape",
			},
			expectedFunc: "KeyTap",
		},
		{
			name: "Type",
			event: KeyboardEvent{
				Action: KeyType,
				Text:   "Hello, world!",
			},
			expectedFunc: "TypeString",
		},
		{
			name: "Combination",
			event: KeyboardEvent{
				Action: KeyCombination,
				Keys:   []string{"control", "c"},
			},
			expectedFunc: "KeyTap",
		},
		{
			name: "Invalid Action",
			event: KeyboardEvent{
				Action: KeyboardAction("invalid"),
			},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mocks for each test case
			ResetMocks()

			// Execute the keyboard event
			err := ExecuteKeyboardEvent(tc.event)

			// Check for expected error
			if tc.expectedError && err == nil {
				t.Error("ExecuteKeyboardEvent() did not return an error for invalid action")
				return
			}

			if !tc.expectedError && err != nil {
				t.Errorf("ExecuteKeyboardEvent() returned an error: %v", err)
				return
			}

			// Skip further checks for error cases
			if tc.expectedError {
				return
			}

			// Check that the expected function was called
			switch tc.expectedFunc {
			case "KeyTap":
				if !GetMockCallCount("KeyTap") {
					t.Error("ExecuteKeyboardEvent() did not call robotgo.KeyTap()")
				}
				args := GetMockLastArgs("KeyTap")
				if tc.event.Action == KeyPress && args["key"] != tc.event.Key {
					t.Errorf("ExecuteKeyboardEvent() called KeyTap with wrong key: got %s, want %s",
						args["key"], tc.event.Key)
				}
			case "TypeString":
				if !GetMockCallCount("TypeString") {
					t.Error("ExecuteKeyboardEvent() did not call robotgo.TypeString()")
				}
				args := GetMockLastArgs("TypeString")
				if args["text"] != tc.event.Text {
					t.Errorf("ExecuteKeyboardEvent() called TypeString with wrong text: got %s, want %s",
						args["text"], tc.event.Text)
				}
			}
		})
	}
}

// TestErrorHandling tests error handling in the helper functions
func TestErrorHandling(t *testing.T) {
	// Test error in ExecuteMouseEvent
	t.Run("ExecuteMouseEvent Error", func(t *testing.T) {
		// Save original function and restore it after the test
		originalMoveMouseFunc := robotgoMoveMouseFunc
		defer func() { robotgoMoveMouseFunc = originalMoveMouseFunc }()

		// Reset mocks
		ResetMocks()

		// Set up mock to return an error
		expectedErr := errors.New("mock error")
		SetMockError("MoveMouse", expectedErr)

		// Execute the mouse event
		event := MouseEvent{
			Action: MouseMove,
			X:      100,
			Y:      200,
		}

		// The error should be propagated
		err := ExecuteMouseEvent(event)
		if err == nil {
			t.Error("ExecuteMouseEvent() did not return an error when the underlying function failed")
		}
	})

	// Test error in ExecuteKeyboardEvent
	t.Run("ExecuteKeyboardEvent Error", func(t *testing.T) {
		// Save original function and restore it after the test
		originalKeyTapFunc := robotgoKeyTapFunc
		defer func() { robotgoKeyTapFunc = originalKeyTapFunc }()

		// Reset mocks
		ResetMocks()

		// Set up mock to return an error
		expectedErr := errors.New("mock error")
		SetMockError("KeyTap", expectedErr)

		// Execute the keyboard event
		event := KeyboardEvent{
			Action: KeyPress,
			Key:    "escape",
		}

		// The error should be propagated
		err := ExecuteKeyboardEvent(event)
		if err == nil {
			t.Error("ExecuteKeyboardEvent() did not return an error when the underlying function failed")
		}
	})
}

// TestHelperFunctions tests the helper functions
func TestHelperFunctions(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Test GetScreenSize
	t.Run("GetScreenSize", func(t *testing.T) {
		width, height := GetScreenSize()
		if width <= 0 || height <= 0 {
			t.Errorf("GetScreenSize() returned invalid dimensions: (%d, %d)", width, height)
		}
		t.Logf("Screen size: %dx%d", width, height)
	})

	// Test GetMousePosition
	t.Run("GetMousePosition", func(t *testing.T) {
		x, y := GetMousePosition()
		// We can't assert exact values since the mouse position can change
		// but we can check that the values are within reasonable bounds
		width, height := GetScreenSize()
		if x < 0 || x > width || y < 0 || y > height {
			t.Errorf("GetMousePosition() returned position outside screen bounds: (%d, %d), screen: %dx%d", x, y, width, height)
		}
		t.Logf("Mouse position: (%d, %d)", x, y)
	})
}

// TestMouseEventHelpers tests the mouse event helper functions
func TestMouseEventHelpers(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Skip on CI or non-interactive environments
	if testing.Short() {
		t.Skip("Skipping mouse event tests in short mode")
	}

	// Get initial mouse position
	initialX, initialY := GetMousePosition()
	t.Logf("Initial mouse position: (%d, %d)", initialX, initialY)

	// Test moving the mouse
	t.Run("MouseMove", func(t *testing.T) {
		// Move the mouse to a specific position
		targetX := initialX + 10
		targetY := initialY + 10

		// Create a mouse event
		event := MouseEvent{
			Action: MouseMove,
			X:      targetX,
			Y:      targetY,
		}

		// Execute the event
		err := executeMouseMove(event.X, event.Y)
		if err != nil {
			t.Errorf("executeMouseMove() returned an error: %v", err)
		}

		// Get the new mouse position
		newX, newY := GetMousePosition()
		t.Logf("New mouse position: (%d, %d)", newX, newY)

		// Move the mouse back to the initial position
		err = executeMouseMove(initialX, initialY)
		if err != nil {
			t.Errorf("executeMouseMove() returned an error when moving back: %v", err)
		}
	})
}

// TestKeyboardEventHelpers tests the keyboard event helper functions
func TestKeyboardEventHelpers(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Skip on CI or non-interactive environments
	if testing.Short() {
		t.Skip("Skipping keyboard event tests in short mode")
	}

	// Test typing text
	t.Run("KeyboardType", func(t *testing.T) {
		// Create a keyboard event
		event := KeyboardEvent{
			Action: KeyboardType,
			Text:   "test",
		}

		// Execute the event
		err := executeKeyboardType(event.Text)
		if err != nil {
			t.Errorf("executeKeyboardType() returned an error: %v", err)
		}
	})

	// Test key press
	t.Run("KeyboardPress", func(t *testing.T) {
		// Create a keyboard event
		event := KeyboardEvent{
			Action: KeyboardPress,
			Key:    "escape",
		}

		// Execute the event
		err := executeKeyboardPress(event.Key, event.Modifiers)
		if err != nil {
			t.Errorf("executeKeyboardPress() returned an error: %v", err)
		}
	})
}
