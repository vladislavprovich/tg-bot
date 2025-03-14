package handler

import (
	"context"

	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vladislavprovich/TG-bot/internal/keyboard"
)

const (
	// Frequency of starting and returning the channel to update in seconds. Default 60.
	updateDelay = 60
	// Offset is the last Update ID to include. Default 0.
	offSet = 0
)

func ProcessUpdates(ctx context.Context, bot *tgbotapi.BotAPI, buttonHandler *keyboard.HandleButtons, messageHandler *keyboard.HandleButtons, logger *logrus.Logger) {
	u := tgbotapi.NewUpdate(offSet)
	u.Timeout = updateDelay

	updates := bot.GetUpdatesChan(u)
	logger.Info("Processing updates...")
	for {
		select {
		case update := <-updates:
			if update.CallbackQuery != nil {
				go buttonHandler.HandleCallbackQuery(ctx, bot, update)
				if update.CallbackQuery != nil {
					go buttonHandler.DeleteButtons(ctx, bot, update)
				}
			} else if update.Message != nil {
				go messageHandler.HandleMessage(ctx, bot, update)
				if update.Message != nil {
					go messageHandler.UserRegister(ctx, bot, update)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
