package logger

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ConfigLogger struct {
	Level string `envconfig:"LOG_LEVEL"`
}

func (c *ConfigLogger) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Level, validation.Required),
	)
}
