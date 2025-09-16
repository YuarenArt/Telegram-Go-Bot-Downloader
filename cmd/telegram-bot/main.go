package main

import (
	"log"
	"youtube_downloader/internal/app"
	"youtube_downloader/internal/config"
)

func main() {

	cfg := config.NewConfig()
	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}
	app.Run()
}
