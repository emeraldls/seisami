package main

import (
	"embed"
	"fmt"
	"net/url"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	_ = godotenv.Load(".env")
	app := NewApp()

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
		Mac: &mac.Options{
			OnUrlOpen: func(url string) {
				handleDeepLink(app, url)

				runtime.EventsEmit(app.ctx, "cloud:setup_started")
				go func() {
					err := app.bootstrapCloud()
					if err != nil {
						runtime.EventsEmit(app.ctx, "cloud:setup_failed", err.Error())
						return
					}
					runtime.EventsEmit(app.ctx, "cloud:setup_success")
				}()
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

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

	if u.Host != "auth" || u.Path != "/callback" {
		return
	}

	query := u.Query()
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
		go func() {
			runtime.EventsEmit(ctx, "auth:desktop_callback", map[string]string{
				"token": token,
				"code":  code,
				"state": state,
			})
		}()
	}
}
