package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Account struct {
	ID           int64              `json:"id"`
	Email        pgtype.Text        `json:"email"`
	PasswordHash pgtype.Text        `json:"password_hash"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

type LinkCode struct {
	Code      string             `json:"code"`
	AccountID int64              `json:"account_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	UsedAt    pgtype.Timestamptz `json:"used_at"`
}

type Session struct {
	ID        string             `json:"id"`
	AccountID int64              `json:"account_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

type TelegramLink struct {
	TelegramID int64              `json:"telegram_id"`
	AccountID  int64              `json:"account_id"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
}
