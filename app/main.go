package main

import (
	"embed"
	"fmt"
	"log"
	"net/url"
	"runtime"

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
	app := NewApp()

	var macOptions *mac.Options
	var windowsOptions *windows.Options

	switch runtime.GOOS {
	case "darwin":
		macOptions = &mac.Options{
			OnUrlOpen: func(url string) {
				handleDeepLink(app, url)

				wailsRuntime.EventsEmit(app.ctx, "cloud:setup_started")
				go func() {
					err := app.bootstrapCloud()
					if err != nil {
						wailsRuntime.EventsEmit(app.ctx, "cloud:setup_failed", err.Error())
						return
					}
					wailsRuntime.EventsEmit(app.ctx, "cloud:setup_success")
				}()
			},
		}

	case "windows":
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "seisami",
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
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

// TODO: update this function to return error
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
	err = app.cloud.InitCloud()

	if err != nil {
		log.Println(err)
	}

	if ctx != nil {
		go func() {
			wailsRuntime.EventsEmit(ctx, "auth:desktop_callback", map[string]string{
				"token": token,
				"code":  code,
				"state": state,
			})
		}()
	}
}
