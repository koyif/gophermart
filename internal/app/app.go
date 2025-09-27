package app

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/koyif/gophermart/internal/config"
)

type App struct {
	Config *config.Config
	DB     *sql.DB
}

func New(cfg *config.Config) (*App, error) {
	dbPool, err := initDB(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &App{
		Config: cfg,
		DB:     dbPool,
	}, nil
}

func initDB(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err = db.Ping(); err != nil {
		err := db.Close()
		if err != nil {
			return nil, fmt.Errorf("error closing database after ping failure: %w", err)
		}
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

func (app App) Run(ctx context.Context) error {

	return nil
}
