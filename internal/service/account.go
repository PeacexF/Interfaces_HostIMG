package service

import (
	"context"
	"errors"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type AccountService struct {
	queries  db.Querier
	sessions *SessionService
}

func NewAccountService(queries db.Querier, sessions *SessionService) *AccountService {
	return &AccountService{queries: queries, sessions: sessions}
}

func (a *AccountService) Signup(ctx context.Context, email, password string) (accountID int64, token string, expiresAt time.Time, err error) {
	emailArg := pgtype.Text{String: email, Valid: true}

	if _, err := a.queries.GetAccountByEmail(ctx, emailArg); err == nil {
		return 0, "", time.Time{}, ErrEmailTaken
	} else if !isNoRows(err) {
		return 0, "", time.Time{}, err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return 0, "", time.Time{}, err
	}

	account, err := a.queries.CreateAccount(ctx, db.CreateAccountParams{
		Email:        emailArg,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, "", time.Time{}, ErrEmailTaken
		}
		return 0, "", time.Time{}, err
	}

	token, expiresAt, err = a.sessions.Create(ctx, account.ID)
	if err != nil {
		return 0, "", time.Time{}, err
	}
	return account.ID, token, expiresAt, nil
}

func (a *AccountService) Login(ctx context.Context, email, password string) (token string, expiresAt time.Time, err error) {
	account, err := a.queries.GetAccountByEmail(ctx, pgtype.Text{String: email, Valid: true})
	if err != nil {
		if isNoRows(err) {
			return "", time.Time{}, ErrInvalidCredentials
		}
		return "", time.Time{}, err
	}

	if !account.PasswordHash.Valid || !verifyPassword(account.PasswordHash.String, password) {
		return "", time.Time{}, ErrInvalidCredentials
	}

	return a.sessions.Create(ctx, account.ID)
}
