package service

import (
	"github.com/vladislavprovich/TG-bot/internal/models"
	"github.com/vladislavprovich/TG-bot/pkg/shortener"
)

type converterToShortener struct {
}

func newConverterToShortener() *converterToShortener {
	return &converterToShortener{}
}

func (c *converterToShortener) ConvertToCreateShortURLRequest(
	req models.CreateShortURLRequest,
) *shortener.CreateShortURLRequest {
	return &shortener.CreateShortURLRequest{
		URL:         req.OriginalURL,
		CustomAlias: req.CustomAlias,
	}
}

func (c *converterToShortener) ConverterToNewStats(
	req models.GetURLStatusRequest,
) *shortener.GetShortURLStatsRequest {
	return &shortener.GetShortURLStatsRequest{
		ShortURL: req.ShortURL,
	}
}
