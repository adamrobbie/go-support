package remote

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/adamrobbie/go-support/pkg/permissions"
	"github.com/go-vgo/robotgo"
)

// MouseAction represents a mouse action type
type MouseAction string

// KeyboardAction represents a keyboard action type
type KeyboardAction string

const (
	// Mouse actions
	MouseMove     MouseAction = "move"
	MouseClick    MouseAction = "click"
	MouseDblClick MouseAction = "doubleClick"
	MouseDrag     MouseAction = "drag"
	MouseScroll   MouseAction = "scroll"
	MouseDown     MouseAction = "down"
	MouseUp       MouseAction = "up"

	// Keyboard actions
	KeyPress       KeyboardAction = "press"
	KeyDown        KeyboardAction = "down"
	KeyUp          KeyboardAction = "up"
	KeyType        KeyboardAction = "type"
	KeyCombination KeyboardAction = "combination"
)

// MouseButton represents a mouse button
type MouseButton string

const (
	LeftButton   MouseButton = "left"
	RightButton  MouseButton = "right"
	MiddleButton MouseButton = "middle"
)

// MouseEvent represents a mouse event
type MouseEvent struct {
	Action MouseAction `json:"action"`
	X      int         `json:"x"`
	Y      int         `json:"y"`
	Button MouseButton `json:"button,omitempty"`
	Double bool        `json:"double,omitempty"`
	Amount int         `json:"amount,omitempty"` // For scrolling
}

// KeyboardEvent represents a keyboard event
type KeyboardEvent struct {
	Action KeyboardAction `json:"action"`
	Key    string         `json:"key"`
	Keys   []string       `json:"keys,omitempty"` // For key combinations
	Text   string         `json:"text,omitempty"` // For typing text
}

// RemoteController handles remote control operations
type RemoteController struct {
	permManager permissions.Manager
	verbose     bool
}

// NewRemoteController creates a new remote controller
func NewRemoteController(permManager permissions.Manager, verbose bool) *RemoteController {
	return &RemoteController{
		permManager: permManager,
		verbose:     verbose,
	}
}

// GetScreenSize returns the screen size
func (rc *RemoteController) GetScreenSize() (int, int, error) {
	// Check permissions first
	if err := rc.checkPermissions(); err != nil {
		return 0, 0, err
	}

	width, height := robotgo.GetScreenSize()
	return width, height, nil
}

// GetMousePosition returns the current mouse position
func (rc *RemoteController) GetMousePosition() (int, int, error) {
	// Check permissions first
	if err := rc.checkPermissions(); err != nil {
		return 0, 0, err
	}

	x, y := robotgo.GetMousePos()
	return x, y, nil
}

// ExecuteMouseEvent executes a mouse event
func (rc *RemoteController) ExecuteMouseEvent(event MouseEvent) error {
	// Check permissions first
	if err := rc.checkPermissions(); err != nil {
		log.Printf("Permission check failed: %v", err)
		return err
	}

	if rc.verbose {
		log.Printf("Executing mouse event: %+v", event)
	}

	switch event.Action {
	case MouseMove:
		log.Printf("Moving mouse to (%d,%d)", event.X, event.Y)

		// Try multiple movement methods in sequence

		// Method 1: Basic Move
		robotgo.Move(event.X, event.Y)

		// Verify position
		x, y := robotgo.GetMousePos()
		if x == event.X && y == event.Y {
			if rc.verbose {
				log.Printf("Basic Move successful, mouse at (%d,%d)", x, y)
			}
			return nil
		}

		if rc.verbose {
			log.Printf("Basic Move failed, mouse at (%d,%d), trying MoveSmooth", x, y)
		}

		// Method 2: MoveSmooth
		robotgo.MoveMouseSmooth(event.X, event.Y, 1.0, 1.0)

		// Verify position
		x, y = robotgo.GetMousePos()
		if x == event.X && y == event.Y {
			if rc.verbose {
				log.Printf("MoveSmooth successful, mouse at (%d,%d)", x, y)
			}
			return nil
		}

		if rc.verbose {
			log.Printf("MoveSmooth failed, mouse at (%d,%d), trying DragSmooth", x, y)
		}

		// Method 3: DragSmooth
		robotgo.DragSmooth(event.X, event.Y)

		// Verify position
		x, y = robotgo.GetMousePos()
		if x == event.X && y == event.Y {
			if rc.verbose {
				log.Printf("DragSmooth successful, mouse at (%d,%d)", x, y)
			}
			return nil
		}

		// Method 4: macOS-specific AppleScript fallback (only on macOS)
		if runtime.GOOS == "darwin" {
			if rc.verbose {
				log.Printf("All RobotGo methods failed, trying macOS-specific fallback")
			}

			err := macOSMoveMouse(event.X, event.Y, rc.verbose)
			if err != nil {
				if rc.verbose {
					log.Printf("macOS fallback failed: %v", err)
				}
			} else {
				// Verify position
				x, y = robotgo.GetMousePos()
				if x == event.X && y == event.Y {
					if rc.verbose {
						log.Printf("macOS fallback successful, mouse at (%d,%d)", x, y)
					}
					return nil
				}
			}
		}

		if rc.verbose {
			log.Printf("All movement methods failed, mouse at (%d,%d)", x, y)
			log.Printf("This may indicate a permissions issue or a problem with RobotGo")
		}

		// Continue even if movement failed

	case MouseClick:
		button := "left"
		if event.Button == RightButton {
			button = "right"
		} else if event.Button == MiddleButton {
			button = "center"
		}

		if event.X > 0 || event.Y > 0 {
			// Move to position first
			err := rc.ExecuteMouseEvent(MouseEvent{
				Action: MouseMove,
				X:      event.X,
				Y:      event.Y,
			})
			if err != nil {
				return fmt.Errorf("failed to move mouse before click: %w", err)
			}
		}

		// Try RobotGo click
		robotgo.MouseClick(button, event.Double)

		// If on macOS, try fallback if needed
		if runtime.GOOS == "darwin" && rc.verbose {
			// We don't have a way to verify if the click worked, so just try the fallback
			// if verbose mode is enabled (assuming this is for debugging)
			macOSClickMouse(button, event.Double, rc.verbose)
		}

	case MouseDblClick:
		// Reuse the click handler with Double=true
		return rc.ExecuteMouseEvent(MouseEvent{
			Action: MouseClick,
			X:      event.X,
			Y:      event.Y,
			Button: event.Button,
			Double: true,
		})

	case MouseDown:
		button := "left"
		if event.Button == RightButton {
			button = "right"
		} else if event.Button == MiddleButton {
			button = "center"
		}

		if event.X > 0 || event.Y > 0 {
			// Move to position first
			err := rc.ExecuteMouseEvent(MouseEvent{
				Action: MouseMove,
				X:      event.X,
				Y:      event.Y,
			})
			if err != nil {
				return fmt.Errorf("failed to move mouse before down: %w", err)
			}
		}

		// Try RobotGo toggle
		robotgo.Toggle(button)

		// If on macOS, try fallback if needed
		if runtime.GOOS == "darwin" && rc.verbose {
			// We don't have a way to verify if the toggle worked, so just try the fallback
			// if verbose mode is enabled (assuming this is for debugging)
			macOSToggleMouse(button, "down", rc.verbose)
		}

	case MouseUp:
		button := "left"
		if event.Button == RightButton {
			button = "right"
		} else if event.Button == MiddleButton {
			button = "center"
		}

		if event.X > 0 || event.Y > 0 {
			// Move to position first
			err := rc.ExecuteMouseEvent(MouseEvent{
				Action: MouseMove,
				X:      event.X,
				Y:      event.Y,
			})
			if err != nil {
				return fmt.Errorf("failed to move mouse before up: %w", err)
			}
		}

		// Try RobotGo toggle
		robotgo.Toggle(button, "up")

		// If on macOS, try fallback if needed
		if runtime.GOOS == "darwin" && rc.verbose {
			// We don't have a way to verify if the toggle worked, so just try the fallback
			// if verbose mode is enabled (assuming this is for debugging)
			macOSToggleMouse(button, "up", rc.verbose)
		}

	case MouseDrag:
		// Get current position
		startX, startY, err := rc.GetMousePosition()
		if err != nil {
			return fmt.Errorf("failed to get mouse position: %w", err)
		}

		// Press mouse button down
		err = rc.ExecuteMouseEvent(MouseEvent{
			Action: MouseDown,
			Button: event.Button,
		})
		if err != nil {
			return fmt.Errorf("failed to press mouse button: %w", err)
		}

		// Move to target position
		err = rc.ExecuteMouseEvent(MouseEvent{
			Action: MouseMove,
			X:      event.X,
			Y:      event.Y,
		})
		if err != nil {
			// Release mouse button before returning error
			rc.ExecuteMouseEvent(MouseEvent{
				Action: MouseUp,
				Button: event.Button,
			})
			return fmt.Errorf("failed to move mouse during drag: %w", err)
		}

		// Small delay to ensure the drag is registered
		time.Sleep(50 * time.Millisecond)

		// Release mouse button
		err = rc.ExecuteMouseEvent(MouseEvent{
			Action: MouseUp,
			Button: event.Button,
		})
		if err != nil {
			return fmt.Errorf("failed to release mouse button: %w", err)
		}

		if rc.verbose {
			log.Printf("Dragged from (%d,%d) to (%d,%d)", startX, startY, event.X, event.Y)
		}

	case MouseScroll:
		// Use Scroll for mouse scrolling
		robotgo.Scroll(0, event.Amount)

	default:
		return fmt.Errorf("unknown mouse action: %s", event.Action)
	}

	return nil
}

// ExecuteKeyboardEvent executes a keyboard event
func (rc *RemoteController) ExecuteKeyboardEvent(event KeyboardEvent) error {
	// Check permissions first
	if err := rc.checkPermissions(); err != nil {
		return err
	}

	if rc.verbose {
		log.Printf("Executing keyboard event: %+v", event)
	}

	switch event.Action {
	case KeyPress:
		// Try RobotGo first
		robotgo.KeyTap(event.Key)

		// If on macOS, try fallback if needed
		if runtime.GOOS == "darwin" && rc.verbose {
			// We don't have a way to verify if the key tap worked, so just try the fallback
			// if verbose mode is enabled (assuming this is for debugging)
			macOSKeyTap(event.Key, rc.verbose)
		}

	case KeyDown:
		robotgo.KeyToggle(event.Key, "down")

	case KeyUp:
		robotgo.KeyToggle(event.Key, "up")

	case KeyType:
		// Try RobotGo first
		robotgo.TypeStr(event.Text)

		// If on macOS, try fallback if needed
		if runtime.GOOS == "darwin" && rc.verbose {
			// We don't have a way to verify if the typing worked, so just try the fallback
			// if verbose mode is enabled (assuming this is for debugging)
			macOSTypeText(event.Text, rc.verbose)
		}

	case KeyCombination:
		if len(event.Keys) > 0 {
			// Last element is the key to tap
			key := event.Keys[len(event.Keys)-1]
			// All other elements are modifiers
			modifiers := event.Keys[:len(event.Keys)-1]

			robotgo.KeyTap(key, modifiers)
		} else {
			return fmt.Errorf("key combination requires at least one key")
		}

	default:
		return fmt.Errorf("unknown keyboard action: %s", event.Action)
	}

	return nil
}

// checkPermissions checks if the remote control permission is granted
func (rc *RemoteController) checkPermissions() error {
	if rc.permManager == nil {
		// If no permission manager is provided, assume permissions are granted
		return nil
	}

	// Use the EnsurePermission method to check and request permission if needed
	granted, err := rc.permManager.EnsurePermission(permissions.RemoteControl)
	if err != nil {
		return fmt.Errorf("failed to check remote control permission: %w", err)
	}

	if !granted {
		return fmt.Errorf("remote control permission not granted")
	}

	return nil
}

// Standalone functions for backward compatibility

// GetScreenSize returns the screen size
func GetScreenSize() (int, int) {
	width, height := robotgo.GetScreenSize()
	return width, height
}

// GetMousePosition returns the current mouse position
func GetMousePosition() (int, int) {
	return robotgo.GetMousePos()
}

// ExecuteMouseEvent executes a mouse event
func ExecuteMouseEvent(event MouseEvent) error {
	// Create a default controller without permission checks for backward compatibility
	controller := &RemoteController{
		verbose: false,
	}
	return controller.ExecuteMouseEvent(event)
}

// ExecuteKeyboardEvent executes a keyboard event
func ExecuteKeyboardEvent(event KeyboardEvent) error {
	// Create a default controller without permission checks for backward compatibility
	controller := &RemoteController{
		verbose: false,
	}
	return controller.ExecuteKeyboardEvent(event)
}
