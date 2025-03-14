package service

import (
	"github.com/vladislavprovich/TG-bot/internal/models"
	"github.com/vladislavprovich/TG-bot/internal/repository"
)

type converterToStorage struct {
}

func newConverterToStorage() *converterToStorage {
	return &converterToStorage{}
}

func (c *converterToStorage) ConvertToSaveURLReq(
	req models.CreateShortURLRequest,
	shortURL string,
	userID string,
) *repository.SaveURLRequest {
	return &repository.SaveURLRequest{
		UserID: userID,
		URL: &repository.URLCombined{
			OriginalURL: req.OriginalURL,
			ShortURL:    shortURL,
		},
	}
}

func (c *converterToStorage) converterToTgID(
	tgID int64,
) *repository.GetUserByTgIDRequest {
	return &repository.GetUserByTgIDRequest{
		TgID: tgID,
	}
}

func (c *converterToStorage) converterToNewUser(
	req models.CreateNewUserRequest,
	userID string,
) *repository.SaveUserRequest {
	return &repository.SaveUserRequest{
		TgID:     req.TgID,
		UserID:   userID,
		UserName: req.UserName,
	}
}
