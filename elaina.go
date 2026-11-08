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
	configPath := flag.String("config", "./config/elaina.yaml", "Path to the configuration file")
	flag.Parse()

	if err := config.Init(*configPath); err != nil {
		logrus.WithError(err).Fatal("Failed to init config")
	}

	// Require an administrator password before starting the server.
	appPassword := config.App.Password
	if appPassword == "" || len(appPassword) < 8 {
		logrus.Fatal("app.password must contain at least 8 characters")
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
