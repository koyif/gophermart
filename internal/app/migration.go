package app

import (
	"errors"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/koyif/gophermart/pkg/logger"
)

func RunMigrations(url string) {
	if url == "" {
		logger.Log.Info("no database URL provided, skipping migrations")
		return
	}

	url = strings.ReplaceAll(url, "postgres://", "pgx5://")

	logger.Log.Info("starting migration")

	m, err := migrate.New("file://migrations", url)
	if err != nil {
		log.Fatalf("error creating migration instance: %v", err)
	}

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Log.Info("nothing changed")
		} else {
			log.Fatalf("error running migration: %v", err)
		}
	}

	logger.Log.Info("migration complete")
}
