package main

import (
	"context"
	"errors"
	"github.com/koyif/gophermart/internal/app"
	"github.com/koyif/gophermart/internal/config"
	"github.com/koyif/gophermart/pkg/logger"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	a.Run(ctx)
	ongoingCtx, cancelOngoingRequests := context.WithCancel(context.Background())
	server := &http.Server{
		Addr: ":8080",
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go startServer(a)

	<-ctx.Done()
	logger.Log.Info("shutting down")

	logger.Log.Info("stopping server")
	if err = server.Shutdown(ongoingCtx); err != nil {
		logger.Log.Error("error shutting down server", logger.Error(err))
	}
	logger.Log.Info("server stopped")

	logger.Log.Info("waiting for ongoing requests to finish")
	select {
	case <-ongoingCtx.Done():
	case <-time.After(5 * time.Second):
	}

	cancelOngoingRequests()
	logger.Log.Info("ongoing requests finished")

	logger.Log.Info("closing database connection")
	if err = a.DB.Close(); err != nil {
		logger.Log.Error("error closing database connection", logger.Error(err))
	}
	logger.Log.Info("database connection closed")

	logger.Log.Info("shutdown complete")
}

func startServer(a *app.App) {
	logger.Log.Info("starting server", logger.String("address", a.Config.Addr))
	if err := http.ListenAndServe(a.Config.Addr, a.Router()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Log.Error("server error", logger.Error(err))
	}
}
