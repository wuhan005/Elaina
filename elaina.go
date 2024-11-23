package main

import (
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	flag.Parse()

	if err := config.Init(); err != nil {
		logrus.WithError(err).Fatal("Failed to init config")
	}

	// Check environment config, make sure the application is safe enough.
	appPassword := config.App.Password
	if appPassword == "" || len(appPassword) < 8 {
		logrus.Fatal("APP_PASSWORD is not strong enough")
	}

	dbInstance, err := db.Init()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	r, err := route.New(dbInstance)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create route")
	}
	r.Run(*port)
}
