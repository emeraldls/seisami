package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

var customDBPath string

// SetCustomDBPath sets a custom database path for testing multiple instances
func SetCustomDBPath(path string) {
	customDBPath = path
}

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

		return filepath.Join(home, ".seisami")
	}
}

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

		return filepath.Join(home, "Seisami", "Recordings")
	}
}

func GetDBPath() string {
	if customDBPath != "" {
		return customDBPath
	}
	return filepath.Join(GetAppDataDir(), "seisami.db")
}
