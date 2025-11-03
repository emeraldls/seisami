//go:build darwin

package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Platform-specific initialization for macOS
func (a *App) initPlatformSpecific() {
	fmt.Printf("Initializing %s platform features...\n", platformName)

	// Check accessibility permissions on macOS
	accessibilityStatus := checkAccessibilityPermission()
	if accessibilityStatus != 1 {
		fmt.Println("Accessibility permissions not granted - hotkey monitoring disabled")
		runtime.EventsEmit(a.ctx, "accessibility:permission_denied",
			fmt.Sprintf("Accessibility permissions are required for %s monitoring. Please grant them in System Settings.", hotkeyName))
		return
	}

	// Start the Fn key listener
	go startHotkeyListener()
	go a.handleHotkeyPress()
}

// Platform-specific cleanup for macOS
func (a *App) cleanupPlatformSpecific() {
	// macOS cleanup if needed
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
