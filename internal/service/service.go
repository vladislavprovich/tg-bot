package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/vladislavprovich/TG-bot/pkg"

	"github.com/sirupsen/logrus"
	"github.com/vladislavprovich/TG-bot/internal/models"
	"github.com/vladislavprovich/TG-bot/internal/repository"

	"github.com/vladislavprovich/TG-bot/pkg/shortener"
)

type URLService interface {
	CreateShortURL(ctx context.Context, req models.CreateShortURLRequest) (models.CreateShortURLResponse, error)
	GetListURL(ctx context.Context, id models.GetListRequest) ([]*models.GetListResponse, error)
	DeleteShortURL(ctx context.Context, url models.DeleteShortURL) error
	DeleteAllURL(ctx context.Context, id models.DeleteAllURL) error
	CreateUserByTgID(ctx context.Context, req models.CreateNewUserRequest) (string, error)
	GetURLStatus(ctx context.Context, req models.GetURLStatusRequest) ([]*models.GetURLStatusResponse, error)
}

type (
	Service struct {
		client             shortener.Client
		repo               repository.URLRepository
		repoUser           repository.UserRepository
		logger             *logrus.Logger
		convertToShortener *converterToShortener
		converterToStorage *converterToStorage
	}

	Params struct {
		Repo     repository.URLRepository
		RepoUser repository.UserRepository
		Logger   *logrus.Logger
		Client   shortener.Client
	}
)

func NewService(params Params) URLService {
	return &Service{
		client:             params.Client,
		repo:               params.Repo,
		repoUser:           params.RepoUser,
		logger:             params.Logger,
		convertToShortener: newConverterToShortener(),
		converterToStorage: newConverterToStorage(),
	}
}

func (s *Service) CreateShortURL(
	ctx context.Context,
	req models.CreateShortURLRequest,
) (models.CreateShortURLResponse, error) {
	if req.OriginalURL == "" {
		return models.CreateShortURLResponse{}, errors.New("origin url is empty")
	}
	if req.UserID == "" {
		newReq := s.converterToStorage.converterToTgID(req.TgID)
		userIDresp, err := s.repoUser.GetUserByTgID(ctx, newReq)
		if err != nil {
			return models.CreateShortURLResponse{}, err
		}
		s.logger.Infof("user id check %s", userIDresp.User.UserID)
		req.UserID = userIDresp.User.UserID
	}
	existingUrls, err := s.repo.GetListURL(ctx, &repository.GetListURLRequest{UserID: req.UserID})
	if err != nil {
		s.logger.Errorf("failed to check existing URLs: %v", err)
		return models.CreateShortURLResponse{}, err
	}

	for _, url := range existingUrls {
		if url.OriginalURL == req.OriginalURL {
			return models.CreateShortURLResponse{ShortURL: url.ShortURL}, nil
		}
	}

	convertedShortUrlReq := s.convertToShortener.ConvertToCreateShortURLRequest(req)
	shortURLResp, err := s.client.CreateShortURL(ctx, convertedShortUrlReq)
	if err != nil {
		s.logger.Errorf("failed to create short url: %v", err)
		return models.CreateShortURLResponse{}, err
	}

	saveReq := s.converterToStorage.ConvertToSaveURLReq(req, shortURLResp.ShortURL, req.UserID)
	if err = s.repo.SaveURL(ctx, saveReq); err != nil {
		s.logger.Errorf("failed to save URL: %v", err)
		return models.CreateShortURLResponse{}, err
	}

	return models.CreateShortURLResponse{ShortURL: shortURLResp.ShortURL}, nil
}

func (s *Service) GetListURL(ctx context.Context, id models.GetListRequest) ([]*models.GetListResponse, error) {
	userID, _ := s.repoUser.GetUserByTgID(ctx, &repository.GetUserByTgIDRequest{TgID: id.TgID})
	userID.User.UserID = id.UserID

	urlsRes, err := s.repo.GetListURL(ctx, &repository.GetListURLRequest{UserID: id.UserID})
	if err != nil {
		s.logger.Errorf("failed to get list of URLs: %v", err)
		return nil, err
	}

	var response []*models.GetListResponse
	for _, url := range urlsRes {
		response = append(response, &models.GetListResponse{
			OriginalURL: url.OriginalURL,
			ShortURL:    url.ShortURL,
		})
	}
	return response, nil
}

func (s *Service) DeleteShortURL(ctx context.Context, url models.DeleteShortURL) error {
	userID, _ := s.repoUser.GetUserByTgID(ctx, &repository.GetUserByTgIDRequest{TgID: url.TgID})
	userID.User.UserID = url.UserID
	err := s.repo.DeleteURL(ctx, &repository.DeleteURLRequest{
		UserID:      url.UserID,
		OriginalURL: url.OriginalURL,
		ShortURL:    url.ShortURL,
	})
	if err != nil {
		s.logger.Errorf("failed to delete short URL: %v", err)
		return err
	}
	return nil
}

func (s *Service) DeleteAllURL(ctx context.Context, id models.DeleteAllURL) error {
	userID, _ := s.repoUser.GetUserByTgID(ctx, &repository.GetUserByTgIDRequest{TgID: id.TgID})
	userID.User.UserID = id.UserID

	err := s.repo.DeleteAllURL(ctx, &repository.DeleteAllURLRequest{
		TgID:   id.TgID,
		UserID: id.UserID,
	})
	if err != nil {
		s.logger.Errorf("failed to delete all URLs: %v", err)
	}
	return nil
}

func (s *Service) CreateUserByTgID(ctx context.Context, req models.CreateNewUserRequest) (string, error) {
	reqNew := s.converterToStorage.converterToTgID(req.TgID)

	userResp, err := s.repoUser.GetUserByTgID(ctx, reqNew)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Errorf("error checking user existence: %v", err)
		return "", errors.New("data base Error")
	}

	if userResp != nil {
		return userResp.User.UserID, nil
	}
	userID := uuid.New().String()

	saveUserReq := s.converterToStorage.converterToNewUser(req, userID)
	err = s.repoUser.SaveUser(ctx, saveUserReq)
	if err != nil {
		s.logger.Errorf("failed to save user: %v", err)
		return "", err
	}
	return saveUserReq.UserID, nil
}

func (s *Service) GetURLStatus(ctx context.Context, req models.GetURLStatusRequest) ([]*models.GetURLStatusResponse,
	error) {
	userID, _ := s.repoUser.GetUserByTgID(ctx, &repository.GetUserByTgIDRequest{TgID: req.TgID})
	req.UserID = userID.User.UserID

	urls, err := s.repo.GetListURL(ctx, &repository.GetListURLRequest{TgID: req.TgID, UserID: req.UserID})
	if err != nil {
		s.logger.Errorf("failed to get list of URLs for user ID %s: %v", req.UserID, err)
		return nil, err
	}

	var (
		responses []*models.GetURLStatusResponse
		stats     *shortener.GetShortURLStatsResponse
	)
	for _, url := range urls {
		shortInfoUrls := pkg.ShortInfo(url.ShortURL)
		statsReq := &shortener.GetShortURLStatsRequest{
			ShortURL: shortInfoUrls,
		}
		stats, err = s.client.GetStatsURL(ctx, statsReq)
		if err != nil {
			s.logger.Errorf("failed to get stats for short URL %s: %v", shortInfoUrls, err)
			return nil, err
		}
		responses = append(responses, &models.GetURLStatusResponse{
			ShortURL:      shortInfoUrls,
			RedirectCount: stats.RedirectCount,
			CreatedAt:     stats.CreatedAt,
		})
	}

	if len(responses) == 0 {
		s.logger.Warnf("no valid stats found for user ID %s", req.UserID)
		return nil, errors.New("no valid stats found for user URLs")
	}
	return responses, nil
}
