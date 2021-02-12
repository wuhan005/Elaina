package main

import (
	"os"

	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	_ = log.NewConsole()
	defer log.Stop()

	// Check environment config, make sure the application is safe enough.
	appURL := os.Getenv("APP_URL")
	appPassword := os.Getenv("APP_PASSWORD")

	if appURL == "" {
		log.Fatal("Empty APP_URL")
	}
	if appPassword == "" || len(appPassword) < 8 {
		log.Fatal("APP_PASSWORD is not strong enough")
	}

	err := db.Init()
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}

	r := route.New()
	err = r.Run()
	if err != nil {
		log.Fatal("Failed to start HTTP server: %v", err)
	}
}
