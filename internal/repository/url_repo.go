package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type URLRepository interface {
	SaveURL(ctx context.Context, req *SaveURLRequest) error
	GetListURL(ctx context.Context, req *GetListURLRequest) ([]*URLCombined, error)
	DeleteAllURL(ctx context.Context, req *DeleteAllURLRequest) error
	DeleteURL(ctx context.Context, req *DeleteURLRequest) error
	GetURLStats(ctx context.Context, req *GetURLStatsRequest) ([]*GetURLStatsResponse, error)
}

type urlRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewBotRepository(db *sql.DB, logger *logrus.Logger) URLRepository {
	return &urlRepository{
		db:     db,
		logger: logger,
	}
}

func (r *urlRepository) SaveURL(ctx context.Context, req *SaveURLRequest) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO urls (user_id, original_url, short_url) VALUES ($1, $2, $3)`,
		req.UserID, req.URL.OriginalURL, req.URL.ShortURL)
	if err != nil {
		return err
	}
	return nil
}

func (r *urlRepository) GetListURL(ctx context.Context, req *GetListURLRequest) ([]*URLCombined, error) {
	query := `SELECT original_url, short_url FROM urls WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, req.UserID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err = rows.Close(); err != nil {
			r.logger.Errorf("rows.Close(): %v", err)
		}
	}()

	var urls []*URLCombined
	for rows.Next() {
		var url URLCombined
		if err = rows.Scan(&url.OriginalURL, &url.ShortURL); err != nil {
			return nil, err
		}
		urls = append(urls, &url)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("rows.Err(): %v", err)
		return nil, err
	}

	return urls, nil
}

func (r *urlRepository) DeleteAllURL(ctx context.Context, req *DeleteAllURLRequest) error {
	query := `DELETE FROM urls WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, req.UserID)
	if err != nil {
		r.logger.Errorf("Failed to delete URLs for user %s: %v", req.UserID, err)
		return err
	}
	return nil
}

func (r *urlRepository) DeleteURL(ctx context.Context, req *DeleteURLRequest) error {
	query := `DELETE FROM urls WHERE (user_id = $1 AND short_url = $2)`
	_, err := r.db.ExecContext(ctx, query, req.UserID, req.ShortURL)
	if err != nil {
		r.logger.Errorf("Failed to delete URLs for user %s: %v", req.UserID, err)
		return err
	}
	return nil
}

func (r *urlRepository) GetURLStats(ctx context.Context, req *GetURLStatsRequest) ([]*GetURLStatsResponse, error) {
	query := `SELECT short_url, created_at, redirect_count FROM urls WHERE user_id = $1 AND  short_url = $2`
	rows, err := r.db.QueryContext(ctx, query, req.UserID, req.ShortURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query URL stats: %w", err)
	}
	defer func() {
		if err = rows.Close(); err != nil {
			r.logger.Errorf("rows.Close(): %v", err)
		}
	}()

	var stats []*GetURLStatsResponse
	for rows.Next() {
		var stat GetURLStatsResponse
		if err = rows.Scan(&stat.ShortURL, &stat.CreatedAt, &stat.RedirectCount); err != nil {
			return nil, fmt.Errorf("failed to scan URL stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("rows.Err(): %v", err)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}
