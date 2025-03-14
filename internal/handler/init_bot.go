package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func BotInit(cfg BotConfig) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		panic(err)
	}
	bot.Debug = true
	return bot
}
