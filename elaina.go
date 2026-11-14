package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	if _, err := db.Init(); err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	r, err := route.New()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create route")
	}

	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", *port),
		Handler:           r,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	go func() {
		logrus.WithField("addr", server.Addr).Info("Listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	<-ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		logrus.WithError(err).Error("Failed to shut down server")
	}
}
