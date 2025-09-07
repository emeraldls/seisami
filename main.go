package main

import (
	"embed"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("unable to load environment variables: ", err)
	}

	var OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")

	if OPENAI_API_KEY == "" {
		log.Fatal("API KEY IS MISSING")
	}

	// bars, err := getAudioWaveForm("./audio.wav", 100)
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// fmt.Println("Waveform bars:", bars)

	// return

	// Create an instance of the app structure
	app := NewApp(OPENAI_API_KEY)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "seisami",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
