package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	flag.Parse()

	// Check environment config, make sure the application is safe enough.
	appPassword := os.Getenv("APP_PASSWORD")
	if appPassword == "" || len(appPassword) < 8 {
		logrus.Fatal("APP_PASSWORD is not strong enough")
	}

	dbInstance, err := db.Init()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	r := route.New(dbInstance)
	r.Run(*port)
}
