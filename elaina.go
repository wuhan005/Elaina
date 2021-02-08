package main

import (
	log "unknwon.dev/clog/v2"

	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	_ = log.NewConsole()
	defer log.Stop()

	r := route.New()
	err := r.Run()
	if err != nil {
		log.Fatal("Failed to start HTTP server: %v", err)
	}
}
