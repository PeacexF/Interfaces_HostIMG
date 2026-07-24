package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
)

const sessionTokenBytes = 32

type SessionService struct {
	queries db.Querier
	ttl     time.Duration
}

func NewSessionService(queries db.Querier, ttl time.Duration) *SessionService {
	return &SessionService{queries: queries, ttl: ttl}
}

func (s *SessionService) Create(ctx context.Context, accountID int64) (token string, expiresAt time.Time, err error) {
	token, err = generateSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(s.ttl)

	if err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		ID:        token,
		AccountID: accountID,
		ExpiresAt: pgtypeTimestamptz(expiresAt),
	}); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (db.Account, error) {
	row, err := s.queries.GetSessionWithAccount(ctx, token)
	if err != nil {
		if isNoRows(err) {
			return db.Account{}, ErrNotFound
		}
		return db.Account{}, err
	}

	if row.SessionExpiresAt.Valid && row.SessionExpiresAt.Time.Before(time.Now()) {
		return db.Account{}, ErrNotFound
	}

	return db.Account{
		ID:           row.AccountID,
		Email:        row.AccountEmail,
		PasswordHash: row.AccountPasswordHash,
		CreatedAt:    row.AccountCreatedAt,
	}, nil
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	return s.queries.DeleteSession(ctx, token)
}

func generateSessionToken() (string, error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
