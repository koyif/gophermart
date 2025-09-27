package main

import (
	"context"
	"errors"
	"github.com/koyif/gophermart/internal/app"
	"github.com/koyif/gophermart/internal/config"
	"github.com/koyif/gophermart/pkg/logger"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	log.Printf("loaded config: %+v", cfg)

	if err = logger.Initialize(); err != nil {
		log.Fatalf("error starting logger: %v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		logger.Log.Fatal("error creating app", logger.Error(err))
	}

	app.RunMigrations(cfg.DatabaseURL)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := a.Run(ctx); err != nil {
		logger.Log.Fatal("error running server", logger.Error(err))
	}

	go startServer(a)

	<-ctx.Done()
	logger.Log.Info("shutting down")
}

func startServer(a *app.App) {
	logger.Log.Info("starting server", logger.String("address", a.Config.Addr))
	if err := http.ListenAndServe(a.Config.Addr, a.Router()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Error("server error", logger.Error(err))
	}
}
