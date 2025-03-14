package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

type UserRepository interface {
	SaveUser(ctx context.Context, req *SaveUserRequest) error
	GetUserByTgID(ctx context.Context, req *GetUserByTgIDRequest) (*GetUserByTgIDResponse, error)
}

type userRepo struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewUserRepository(db *sql.DB, logger *logrus.Logger) UserRepository {
	return &userRepo{
		db:     db,
		logger: logger,
	}
}

func (u *userRepo) SaveUser(ctx context.Context, req *SaveUserRequest) error {
	query := `
        INSERT INTO users (user_id, telegram_id, username)
        VALUES ($1, $2, $3)
        ON CONFLICT (telegram_id) DO NOTHING
    `

	_, err := u.db.ExecContext(ctx, query, req.UserID, req.TgID, req.UserName)
	if err != nil {
		u.logger.Errorf("Failed to save user: %v", err)
		return fmt.Errorf("error saving user: %w", err)
	}
	return nil
}

func (u *userRepo) GetUserByTgID(ctx context.Context, req *GetUserByTgIDRequest) (*GetUserByTgIDResponse, error) {
	query := `
        SELECT user_id, telegram_id
        FROM users
        WHERE telegram_id = $1
    `
	var user User
	err := u.db.QueryRowContext(ctx, query, req.TgID).Scan(&user.UserID, &user.TgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.logger.Warnf("User not found: TelegramID %d", req.TgID)
			return nil, err
		}
		u.logger.Errorf("Failed to get user by Telegram ID: %v", err)
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return &GetUserByTgIDResponse{User: &user}, nil
}
