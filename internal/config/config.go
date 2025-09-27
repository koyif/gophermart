package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Addr                 string   `env:"RUN_ADDRESS" env-default:"localhost:8081"`
	AccrualSystemAddress string   `env:"ACCRUAL_SYSTEM_ADDRESS" env-default:"localhost:8080"`
	DatabaseURL          string   `env:"DATABASE_URI"`
	PrivateKey           string   `env:"PRIVATE_KEY" env-default:"privatekey"`
	AuthDisabledURLs     []string `env:"AUTH_DISABLED_URLS" env-default:"/login,/register" env-separator:","`
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Addr, "a", "localhost:8081", "адрес эндпоинта HTTP-сервера")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "localhost:8080", "адрес системы расчёта начислений")
	flag.StringVar(&cfg.DatabaseURL, "d", "", "URL базы данных")

	flag.Parse()

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't read environment variables: %w", err)
	}

	return cfg, nil
}
