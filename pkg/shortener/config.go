package shortener

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Port    string `envconfig:"SERVER_PORT" `
	BaseURL string `envconfig:"BASE_URL" `
}

func (c Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.BaseURL, validation.Required),
	)
}
