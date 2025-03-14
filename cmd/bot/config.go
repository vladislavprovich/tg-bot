package main

import (
	"context"
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/vladislavprovich/TG-bot/internal/handler"
	"github.com/vladislavprovich/TG-bot/internal/repository/postgres"
	"github.com/vladislavprovich/TG-bot/pkg/logger"
	"github.com/vladislavprovich/TG-bot/pkg/shortener"
)

type Config struct {
	Bot      handler.BotConfig
	Database postgres.Config
	Logger   logger.ConfigLogger
	Client   *shortener.Config
}

func LoadConfig(ctx context.Context) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("error loading .env file")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Bot),
		validation.Field(&c.Database),
		validation.Field(&c.Logger),
		validation.Field(&c.Client),
	)
}
