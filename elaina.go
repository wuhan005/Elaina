package main

import (
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	_ = log.NewConsole()
	defer log.Stop()

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
