package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/wuhan005/Elaina/internal/config"
	"github.com/wuhan005/Elaina/internal/db"
	"github.com/wuhan005/Elaina/internal/route"
	"github.com/wuhan005/Elaina/internal/runtime"
)

func main() {
	port := flag.Int("port", 8080, "Web service port")
	configPath := flag.String("config", "./config/elaina.yaml", "Path to the configuration file")
	flag.Parse()

	if err := config.Init(*configPath); err != nil {
		logrus.WithError(err).Fatal("Failed to init config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtime.Init(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to initialize WASM runtime")
	}
	defer func() { _ = runtime.Close(context.Background()) }()

	dbInstance, err := db.Init()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	r, err := route.New(dbInstance)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create route")
	}

	go func() {
		<-ctx.Done()
		r.Stop()
	}()
	r.Run(*port)
}
