package service

import (
	"errors"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TxQuerier interface {
	db.Querier
	WithTx(tx pgx.Tx) *db.Queries
}

var (
	ErrNotFound                = errors.New("not found")
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrEmailTaken              = errors.New("email is already in use")
	ErrLinkCodeInvalid         = errors.New("link code is invalid, expired, or already used")
	ErrAccountsBothHaveHistory = errors.New("both accounts already have independent history and cannot be automatically merged")
)

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func pgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
