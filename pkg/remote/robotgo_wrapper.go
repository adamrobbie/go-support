package remote

import (
	"github.com/go-vgo/robotgo"
)

// Wrapper functions for RobotGo to make them easier to mock in tests

var (
	// Screen functions
	robotgoGetScreenSizeFunc = func() (int, int) {
		return robotgo.GetScreenSize()
	}

	// Mouse functions
	robotgoGetMousePosFunc = func() (int, int) {
		return robotgo.GetMousePos()
	}

	robotgoMoveMouseFunc = func(x, y int) {
		robotgo.MoveMouse(x, y)
	}

	robotgoClickFunc = func(button string, double bool) {
		robotgo.Click(button, double)
	}

	robotgoMouseToggleFunc = func(button, direction string) {
		robotgo.Toggle(button, direction)
	}

	// Keyboard functions
	robotgoTypeStrFunc = func(text string) {
		robotgo.TypeStr(text)
	}

	robotgoKeyTapFunc = func(key string, modifiers ...string) {
		// Convert []string to []interface{} for robotgo.KeyTap
		if len(modifiers) > 0 {
			args := make([]interface{}, len(modifiers))
			for i, mod := range modifiers {
				args[i] = mod
			}
			robotgo.KeyTap(key, args...)
		} else {
			robotgo.KeyTap(key)
		}
	}
)
