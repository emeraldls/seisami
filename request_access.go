package main

// CheckMicrophonePermission checks the current microphone permission status
// Returns: 1 = authorized, 0 = denied/restricted, -1 = not determined
func (a *App) CheckMicrophonePermission() int {
	return checkMicrophonePermission()
}

// RequestMicrophonePermission requests microphone permission and returns the result
func (a *App) RequestMicrophonePermission() bool {
	return requestMicrophonePermissionSync()
}

// OpenMicrophoneSettings opens the system settings for microphone permissions
func (a *App) OpenMicrophoneSettings() {
	openMicrophoneSettings()
}

// CheckAccessibilityPermission checks the current accessibility permission status
// Returns: 1 = authorized, 0 = denied/not authorized
func (a *App) CheckAccessibilityPermission() int {
	return checkAccessibilityPermission()
}

// RequestAccessibilityPermission requests accessibility permission and shows system dialog
func (a *App) RequestAccessibilityPermission() {
	requestAccessibilityPermission()
}

// OpenAccessibilitySettings opens the system settings for accessibility permissions
func (a *App) OpenAccessibilitySettings() {
	openAccessibilitySettings()
}
