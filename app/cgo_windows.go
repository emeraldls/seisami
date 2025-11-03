//go:build windows

package main

/*
#cgo CFLAGS: -I./include
#cgo windows CFLAGS: -I${SRCDIR}/portaudio/include
#cgo windows LDFLAGS: ${SRCDIR}/portaudio/lib/windows/libportaudio.a -lwinmm -lole32 -luuid
#include "portaudio.h"
#include "windows_hotkey.c"
*/
import "C"
import (
	"sync/atomic"
)

// Global state for listener
var (
	hotkeyListenerRunning atomic.Bool
)

// Platform-specific implementations for Windows

// Windows uses F13 key (or configurable) instead of Fn key
func isHotkeyPressed() bool {
	return C.get_hotkey_state() != 0
}

func startHotkeyListener() {
	if hotkeyListenerRunning.Load() {
		return
	}
	hotkeyListenerRunning.Store(true)
	go windowsHotkeyListener()
}

// Windows hotkey listener using keyboard hook
func windowsHotkeyListener() {
	C.start_windows_hotkey_listener()
}

func playRecordingSound() {
	// Play Windows system sound
	C.MessageBeep(C.MB_OK)
}

func checkMicrophonePermission() int {
	// Windows 10/11 doesn't require explicit microphone permission
	// for desktop apps (only UWP apps need it)
	// Return 1 (authorized) by default
	return 1
}

func requestMicrophonePermission() {
	// No-op on Windows for desktop apps
}

func isPermissionRequestCompleted() bool {
	return true
}

func getPermissionRequestResult() int {
	return 1
}

func requestMicrophonePermissionSync() bool {
	return true
}

func openMicrophoneSettings() {
	// Open Windows microphone privacy settings
	C.open_windows_microphone_settings()
}

func checkAccessibilityPermission() int {
	// Windows doesn't require accessibility permission for keyboard hooks
	return 1
}

func requestAccessibilityPermission() {
	// No-op on Windows
}

func openAccessibilitySettings() {
	// No-op on Windows
}

// Platform-specific constants
const (
	platformName              = "Windows"
	requiresAccessibilityPerm = false
	hotkeyName                = "F13 Key"
)
