package keyboard

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"github.com/vladislavprovich/TG-bot/internal/models"
	"github.com/vladislavprovich/TG-bot/internal/service"
)

const (
	Start          = "/start"
	CreateShortURL = "create_short_url"
	RandomURL      = "rand_url"
	CustomURL      = "cust_url"
	ListURL        = "list_short_urls"
	Settings       = "settings"
	BackMainMenu   = "back_to_main"
	ClearHistory   = "clear_history"
	ShowURLStatus  = "show_url_stats"
)

var (
	userStates = make(map[int64]*models.UserAction)
	urls       []*models.GetListResponse
)

type HandleButtons struct {
	service service.URLService
	logger  *logrus.Logger
}

func NewHandleButtons(service service.URLService, logger *logrus.Logger) *HandleButtons {
	return &HandleButtons{service: service, logger: logger}
}

func NewMessageHandler(service service.URLService, logger *logrus.Logger) *HandleButtons {
	return &HandleButtons{service: service, logger: logger}
}

func (h *HandleButtons) HandleCallbackQuery(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		h.logger.Info("CallbackQuery is nil")
		return
	}

	query := update.CallbackQuery
	tgID := query.From.ID

	checkUserReq := models.CreateNewUserRequest{
		TgID: tgID,
	}

	userID, err := h.service.CreateUserByTgID(ctx, checkUserReq)
	if err != nil {
		h.logger.Error(err)
	}

	if query.Message == nil {
		h.logger.Info("Message in CallbackQuery is nil")
		return
	}

	switch query.Data {
	case CreateShortURL:
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
			"Choose an action from the menu.")
		msg.ReplyMarkup = CreateURL()
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case RandomURL:
		userStates[tgID] = &models.UserAction{Action: RandomURL}
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Send original URL:")
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case CustomURL:
		userStates[tgID] = &models.UserAction{Action: CustomURL}
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Send original URL:")
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case ListURL:
		urls, err = h.service.GetListURL(ctx, models.GetListRequest{UserID: userID, TgID: tgID})
		if err != nil {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"Error retrieving URL list.")
			msg.ReplyMarkup = BackMenu()
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}
		if len(urls) == 0 {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"You have no short URLs.")
			msg.ReplyMarkup = BackMenu()
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}

		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Your URLs: ")
		msg.ReplyMarkup = CreateURLListWithDeleteButtons(urls)
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case Settings:
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Settings:")
		msg.ReplyMarkup = ClearAndBack()
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case BackMainMenu:
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
			"Choose action in menu.")
		msg.ReplyMarkup = MainMenu()
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}

	case ClearHistory:
		if err = h.service.DeleteAllURL(ctx, models.DeleteAllURL{UserID: userID, TgID: tgID}); err != nil {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"Error clearing history.")
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "History cleared.")
		msg.ReplyMarkup = BackMenu()
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}
	case ShowURLStatus:
		var stats []*models.GetURLStatusResponse
		stats, err = h.service.GetURLStatus(ctx, models.GetURLStatusRequest{UserID: userID, TgID: tgID})
		if err != nil || len(stats) == 0 {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"Error retrieving URL stats.")
			msg.ReplyMarkup = BackMenu()
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}
		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Your URLs stats:")
		msg.ReplyMarkup = CreateURLStatusButton(stats)
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}
	}
}

func (h *HandleButtons) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		h.logger.Info("Message is nil, skipping update.")
		return
	}

	tgID := update.Message.From.ID
	messageText := update.Message.Text
	var userID string

	user, exists := userStates[tgID]
	if !exists {
		userStates[tgID] = &models.UserAction{}
		h.logger.Infof("Created new user state for userID: %d\n", tgID)
		return
	}

	switch user.Action {
	case RandomURL:
		req := models.CreateShortURLRequest{
			TgID:        tgID,
			UserID:      userID,
			OriginalURL: messageText,
		}

		resp, err := h.service.CreateShortURL(ctx, req)
		if err != nil {
			h.logger.Errorf("Failed to create short URL: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
			if _, err = bot.Send(msg); err != nil {
				h.logger.Errorf("Failed to send error message: %v", err)
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your short URL: "+resp.ShortURL)
		msg.ReplyMarkup = BackMenu()
		if _, err = bot.Send(msg); err != nil {
			h.logger.Errorf("Failed to send short URL: %v", err)
		}

		userStates[tgID] = &models.UserAction{}

	case CustomURL:
		if user.OriginalURL == "" {
			user.OriginalURL = messageText

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send custom short URL:")
			if _, err := bot.Send(msg); err != nil {
				h.logger.Errorf("Failed to request custom short URL: %v", err)
			}
		} else {
			user.CustomURL = messageText
			req := models.CreateShortURLRequest{
				TgID:        tgID,
				UserID:      userID,
				OriginalURL: user.OriginalURL,
				CustomAlias: &user.CustomURL,
			}

			resp, err := h.service.CreateShortURL(ctx, req)
			if err != nil {
				h.logger.Errorf("Failed to create custom short URL: %v", err)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
				if _, err = bot.Send(msg); err != nil {
					h.logger.Errorf("Failed to send error message: %v", err)
				}
				return
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Your short URL: "+resp.ShortURL)
			msg.ReplyMarkup = BackMenu()
			if _, err = bot.Send(msg); err != nil {
				h.logger.Errorf("Failed to send short URL: %v", err)
			}

			userStates[tgID] = &models.UserAction{}
		}

	default:
		if messageText != Start {
			h.logger.Errorf("Unexpected user action: %v\n", user.Action)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Choose an action from the menu.")
			msg.ReplyMarkup = MainMenu()
			if _, err := bot.Send(msg); err != nil {
				h.logger.Errorf("Failed to send default menu: %v", err)
			}
		}
	}
}

func (h *HandleButtons) DeleteButtons(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	query := update.CallbackQuery
	tgID := query.From.ID

	checkUserReq := models.CreateNewUserRequest{
		TgID: tgID,
	}

	userID, err := h.service.CreateUserByTgID(ctx, checkUserReq)
	if err != nil {
		h.logger.Error(err)
	}

	if query.Message == nil {
		h.logger.Info("Message in CallbackQuery is nil")
		return
	}

	deleteButtonsResp := strings.HasPrefix(query.Data, "delete_short_url:")
	if deleteButtonsResp {
		shortURL := strings.TrimPrefix(query.Data, "delete_short_url:")
		h.logger.Infof("Received callback to delete short URL: %s", shortURL)

		if err = h.service.DeleteShortURL(ctx, models.DeleteShortURL{
			TgID:     tgID,
			UserID:   userID,
			ShortURL: shortURL,
		}); err != nil {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"Failed to delete URL.")
			msg.ReplyMarkup = BackMenu()
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}

		for i := 0; i < len(urls); i++ {
			if urls[i].ShortURL == shortURL {
				urls = append(urls[:i], urls[i+1:]...)
				break
			}
		}

		if len(urls) == 0 {
			msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID,
				"All URLs have been deleted.")
			msg.ReplyMarkup = BackMenu()
			_, err = bot.Send(msg)
			if err != nil {
				h.logger.Error(err)
				return
			}
			return
		}

		msg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, "Your URLs:")
		msg.ReplyMarkup = CreateURLListWithDeleteButtons(urls)
		_, err = bot.Send(msg)
		if err != nil {
			h.logger.Error(err)
			return
		}
		return
	}
}

func (h *HandleButtons) UserRegister(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	tgID := update.Message.From.ID
	messageText := update.Message.Text

	if messageText == Start {
		createUserReq := models.CreateNewUserRequest{
			TgID:     tgID,
			UserName: update.Message.From.UserName,
		}

		var err error
		userID, err := h.service.CreateUserByTgID(ctx, createUserReq)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				h.logger.Errorf("Failed to create user: %v", err)
				h.logger.Errorf("USER ID: %v", userID)
			}
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Welcome to the bot! Choose an action from the menu.")
		msg.ReplyMarkup = MainMenu()
		if _, err = bot.Send(msg); err != nil {
			h.logger.Errorf("Failed to send start message: %v", err)
		}
		return
	}
}
