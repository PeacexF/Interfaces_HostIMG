package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CountTelegramLinksForAccount(ctx context.Context, accountID int64) (int64, error)
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	CreateEmptyAccount(ctx context.Context) (Account, error)
	CreateLinkCode(ctx context.Context, arg CreateLinkCodeParams) error
	CreateSession(ctx context.Context, arg CreateSessionParams) error
	CreateTelegramLink(ctx context.Context, arg CreateTelegramLinkParams) error
	DeleteAccount(ctx context.Context, id int64) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteSession(ctx context.Context, id string) error
	GetAccountByEmail(ctx context.Context, email pgtype.Text) (Account, error)
	GetAccountByID(ctx context.Context, id int64) (Account, error)
	GetLinkCode(ctx context.Context, code string) (LinkCode, error)
	GetSessionWithAccount(ctx context.Context, id string) (GetSessionWithAccountRow, error)
	GetTelegramLink(ctx context.Context, telegramID int64) (TelegramLink, error)
	LockTelegramID(ctx context.Context, telegramID int64) error
	MarkLinkCodeUsed(ctx context.Context, code string) error
	MigrateSessionsToAccount(ctx context.Context, arg MigrateSessionsToAccountParams) error
	UpdateAccountCredentials(ctx context.Context, arg UpdateAccountCredentialsParams) error
}

var _ Querier = (*Queries)(nil)
