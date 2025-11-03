//go:build windows

package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Platform-specific initialization for Windows
func (a *App) initPlatformSpecific() {
	fmt.Printf("Initializing %s platform features...\n", platformName)

	// Start the F13 key listener
	go startHotkeyListener()
	go a.handleHotkeyPress()

	// Notify user about the hotkey
	runtime.EventsEmit(a.ctx, "platform:info",
		fmt.Sprintf("Press and hold %s to start recording", hotkeyName))
}

// Platform-specific cleanup for Windows
func (a *App) cleanupPlatformSpecific() {
	// Windows cleanup if needed
}

// Get platform info for frontend
func (a *App) GetPlatformInfo() map[string]interface{} {
	return map[string]interface{}{
		"platform":                  platformName,
		"hotkeyName":                hotkeyName,
		"requiresAccessibilityPerm": requiresAccessibilityPerm,
		"microphonePermission":      checkMicrophonePermission(),
		"accessibilityPermission":   checkAccessibilityPermission(),
	}
}
