package main

import (
	"flag"
	"os"
	"strconv"

	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	_ = log.NewConsole()
	defer log.Stop()

	port := flag.Int("port", 8080, "Web service port")
	flag.Parse()

	// Check environment config, make sure the application is safe enough.
	appPassword := os.Getenv("APP_PASSWORD")

	if appPassword == "" || len(appPassword) < 8 {
		log.Fatal("APP_PASSWORD is not strong enough")
	}

	err := os.MkdirAll("./volume", 0755)
	if err != nil {
		log.Fatal("Failed to create path ./volume: %v", err)
	}
	log.Trace("Create ./volume succeed!")

	err = db.Init()
	if err != nil {
		log.Fatal("Failed to connect to database: %v", err)
	}

	r := route.New()
	err = r.Run("0.0.0.0:" + strconv.Itoa(*port))
	if err != nil {
		log.Fatal("Failed to start HTTP server: %v", err)
	}
}
