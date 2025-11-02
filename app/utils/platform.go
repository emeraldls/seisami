package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetAppDataDir returns the platform-specific application data directory
func GetAppDataDir() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/Seisami
		return filepath.Join(home, "Library", "Application Support", "Seisami")
	case "windows":
		// Windows: %APPDATA%/Seisami or fallback to %USERPROFILE%/AppData/Roaming/Seisami
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "Seisami")
		}
		return filepath.Join(home, "AppData", "Roaming", "Seisami")
	case "linux":
		// Linux: ~/.config/Seisami
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir != "" {
			return filepath.Join(configDir, "Seisami")
		}
		return filepath.Join(home, ".config", "Seisami")
	default:
		// Fallback for unknown platforms
		return filepath.Join(home, ".seisami")
	}
}

// GetRecordingsDir returns the platform-specific recordings directory
func GetRecordingsDir() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Music/Seisami
		return filepath.Join(home, "Music", "Seisami")
	case "windows":
		// Windows: %USERPROFILE%/Documents/Seisami/Recordings
		return filepath.Join(home, "Documents", "Seisami", "Recordings")
	case "linux":
		// Linux: ~/Documents/Seisami/Recordings
		return filepath.Join(home, "Documents", "Seisami", "Recordings")
	default:
		// Fallback
		return filepath.Join(home, "Seisami", "Recordings")
	}
}

// GetDBPath returns the platform-specific database file path
func GetDBPath() string {
	return filepath.Join(GetAppDataDir(), "seisami.db")
}
