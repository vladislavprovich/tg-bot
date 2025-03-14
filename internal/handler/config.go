package handler

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type BotConfig struct {
	BotToken string `envconfig:"BOT_TOKEN" `
}

func (b *BotConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &b,
		validation.Field(&b.BotToken, validation.Required),
	)
}
