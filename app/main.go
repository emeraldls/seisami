package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"seisami/app/utils"
	"strings"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	_ = godotenv.Load(".env")

	// Check for custom DB path from environment variable (for dev mode) or flag (for production)
	customDBPath := os.Getenv("SEISAMI_DB_PATH")

	// If no env var, try parsing command line flags (for production builds)
	if customDBPath == "" {
		dbPath := flag.String("db", "", "Path to custom database file (e.g., -db=/path/to/custom.db)")
		flag.Parse()
		if *dbPath != "" {
			customDBPath = *dbPath
		}
	}

	// Set custom DB path if provided
	if customDBPath != "" {
		utils.SetCustomDBPath(customDBPath)
		fmt.Printf("Using custom database path: %s\n", customDBPath)
	}

	app := NewApp()
	var macOptions *mac.Options
	var windowsOptions *windows.Options
	var deepLink string

	switch runtime.GOOS {
	case "darwin":
		macOptions = &mac.Options{
			OnUrlOpen: func(url string) {
				handleDeepLink(app, url)
			},
		}

	case "windows":
		if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "seisami://") {
			deepLink = os.Args[1]
		}

		exePath, _ := os.Executable()
		if err := registerWindowsProtocol("seisami", exePath); err != nil {
			fmt.Println("Failed to register protocol:", err)
		}

		windowsOptions = &windows.Options{
			WebviewUserDataPath: "seisami_userdata",
		}
	}

	err := wails.Run(&options.App{
		Title: "seisami",
		// MinWidth:  1024,
		// MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)

			if deepLink != "" {
				go handleDeepLink(app, deepLink)
			}
		},
		Bind: []interface{}{
			app,
		},
		Mac:     macOptions,
		Windows: windowsOptions,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// -------------------------
// Deep link handler
// -------------------------
func handleDeepLink(app *App, deepLink string) {
	if app == nil {
		fmt.Println("App instance not ready for deep link handling")
		return
	}

	ctx := app.ctx
	if ctx == nil {
		fmt.Println("App context not ready; deferring event emission")
	}

	u, err := url.Parse(deepLink)
	if err != nil {
		fmt.Println("Failed to parse deep link:", err)
		return
	}

	query := u.Query()

	if u.Host == "board" && u.Path == "/import" {
		boardID := query.Get("board_id")
		if boardID == "" {
			fmt.Println("Missing board_id in deep link")
			return
		}

		fmt.Println("Board import requested:", boardID)
		if ctx != nil {
			go func() {
				wailsRuntime.EventsEmit(ctx, "board:import_request", map[string]string{
					"board_id": boardID,
				})
			}()
		}
		return
	}

	if u.Host != "auth" || u.Path != "/callback" {
		return
	}

	token := query.Get("token")
	code := query.Get("code")
	state := query.Get("state")

	if token == "" || code == "" || state == "" {
		fmt.Println("Missing required fields in deep link")
		return
	}

	fmt.Println("Auth callback received")
	app.SetLoginToken(token)

	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, "cloud:setup_started")

		go func() {
			if err := app.bootstrapCloud(); err != nil {
				wailsRuntime.EventsEmit(ctx, "cloud:setup_failed", err.Error())
				return
			}
			wailsRuntime.EventsEmit(ctx, "cloud:setup_success")
		}()

		go func() {
			wailsRuntime.EventsEmit(ctx, "auth:desktop_callback", map[string]string{
				"token": token,
				"code":  code,
				"state": state,
			})
		}()
	}
}

// -------------------------
// Windows protocol registration
// -------------------------
func registerWindowsProtocol(appName, exePath string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// seisami:// protocol registration
	cmd := exec.Command("reg", "add", "HKEY_CLASSES_ROOT\\"+appName, "/ve", "/d", "URL:"+strings.Title(appName)+" Protocol", "/f")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("reg", "add", "HKEY_CLASSES_ROOT\\"+appName, "/v", "URL Protocol", "/d", "", "/f")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("reg", "add", "HKEY_CLASSES_ROOT\\"+appName+"\\shell\\open\\command", "/ve",
		"/d", fmt.Sprintf("\"%s\" \"%%1\"", exePath), "/f")
	return cmd.Run()
}
