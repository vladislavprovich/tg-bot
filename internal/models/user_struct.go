package models

import "time"

type User struct {
	Action      string     `json:"action"`
	ID          string     `json:"id"`
	TelegramID  string     `json:"telegram_id"`
	OriginalURL string     `json:"original_url"`
	CustomURL   string     `json:"custom_url"`
	ShortURL    string     `json:"short_url"`
	UserAction  UserAction `json:"user_action"`
}

type UserAction struct {
	Action      string `json:"action"` // "rand_url" or "cust_url"
	OriginalURL string `json:"original_url"`
	CustomURL   string `json:"custom_url"`
	ShortURL    string `json:"short_url"`
}

type CreateShortURLRequest struct {
	TgID        int64   `json:"tg_id"`
	UserID      string  `json:"user_id"`
	OriginalURL string  `json:"original_url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
}

type CreateShortURLResponse struct {
	ShortURL string `json:"short_url"`
}

type DeleteShortURL struct {
	TgID        int64  `json:"tg_id"`
	UserID      string `json:"user_id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type GetListRequest struct {
	TgID   int64  `json:"tg_id"`
	UserID string `json:"user_id"`
}

type GetListResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type DeleteAllURL struct {
	TgID   int64  `json:"tg_id"`
	UserID string `json:"user_id"`
}

type CreateNewUserRequest struct {
	TgID     int64  `json:"tg_id"`
	UserName string `json:"username"`
}

type GetURLStatusRequest struct {
	UserID   string `json:"user_id"`
	TgID     int64  `json:"tg_id"`
	ShortURL string `json:"short_url"`
}

type GetURLStatusResponse struct {
	ShortURL      string    `json:"short_url"`
	RedirectCount int       `json:"redirect_count"`
	CreatedAt     time.Time `json:"created_at"`
}
