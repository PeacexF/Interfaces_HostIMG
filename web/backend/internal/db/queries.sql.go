package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countTelegramLinksForAccount = `-- name: CountTelegramLinksForAccount :one
SELECT COUNT(*) FROM telegram_links WHERE account_id = $1
`

func (q *Queries) CountTelegramLinksForAccount(ctx context.Context, accountID int64) (int64, error) {
	row := q.db.QueryRow(ctx, countTelegramLinksForAccount, accountID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, created_at
`

type CreateAccountParams struct {
	Email        pgtype.Text `json:"email"`
	PasswordHash pgtype.Text `json:"password_hash"`
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRow(ctx, createAccount, arg.Email, arg.PasswordHash)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.CreatedAt,
	)
	return i, err
}

const createEmptyAccount = `-- name: CreateEmptyAccount :one
INSERT INTO accounts DEFAULT VALUES
RETURNING id, email, password_hash, created_at
`

func (q *Queries) CreateEmptyAccount(ctx context.Context) (Account, error) {
	row := q.db.QueryRow(ctx, createEmptyAccount)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.CreatedAt,
	)
	return i, err
}

const createLinkCode = `-- name: CreateLinkCode :exec
INSERT INTO link_codes (code, account_id, expires_at)
VALUES ($1, $2, $3)
`

type CreateLinkCodeParams struct {
	Code      string             `json:"code"`
	AccountID int64              `json:"account_id"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (q *Queries) CreateLinkCode(ctx context.Context, arg CreateLinkCodeParams) error {
	_, err := q.db.Exec(ctx, createLinkCode, arg.Code, arg.AccountID, arg.ExpiresAt)
	return err
}

const createSession = `-- name: CreateSession :exec
INSERT INTO sessions (id, account_id, expires_at)
VALUES ($1, $2, $3)
`

type CreateSessionParams struct {
	ID        string             `json:"id"`
	AccountID int64              `json:"account_id"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.Exec(ctx, createSession, arg.ID, arg.AccountID, arg.ExpiresAt)
	return err
}

const createTelegramLink = `-- name: CreateTelegramLink :exec
INSERT INTO telegram_links (telegram_id, account_id)
VALUES ($1, $2)
`

type CreateTelegramLinkParams struct {
	TelegramID int64 `json:"telegram_id"`
	AccountID  int64 `json:"account_id"`
}

func (q *Queries) CreateTelegramLink(ctx context.Context, arg CreateTelegramLinkParams) error {
	_, err := q.db.Exec(ctx, createTelegramLink, arg.TelegramID, arg.AccountID)
	return err
}

const deleteAccount = `-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1
`

func (q *Queries) DeleteAccount(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, deleteAccount, id)
	return err
}

const deleteExpiredSessions = `-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at < NOW()
`

func (q *Queries) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	result, err := q.db.Exec(ctx, deleteExpiredSessions)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteSession = `-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1
`

func (q *Queries) DeleteSession(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteSession, id)
	return err
}

const getAccountByEmail = `-- name: GetAccountByEmail :one
SELECT id, email, password_hash, created_at FROM accounts WHERE email = $1
`

func (q *Queries) GetAccountByEmail(ctx context.Context, email pgtype.Text) (Account, error) {
	row := q.db.QueryRow(ctx, getAccountByEmail, email)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.CreatedAt,
	)
	return i, err
}

const getAccountByID = `-- name: GetAccountByID :one
SELECT id, email, password_hash, created_at FROM accounts WHERE id = $1
`

func (q *Queries) GetAccountByID(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRow(ctx, getAccountByID, id)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.CreatedAt,
	)
	return i, err
}

const getLinkCode = `-- name: GetLinkCode :one
SELECT code, account_id, created_at, expires_at, used_at FROM link_codes WHERE code = $1
`

func (q *Queries) GetLinkCode(ctx context.Context, code string) (LinkCode, error) {
	row := q.db.QueryRow(ctx, getLinkCode, code)
	var i LinkCode
	err := row.Scan(
		&i.Code,
		&i.AccountID,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UsedAt,
	)
	return i, err
}

const getSessionWithAccount = `-- name: GetSessionWithAccount :one
SELECT
    s.id AS session_id,
    s.expires_at AS session_expires_at,
    a.id AS account_id,
    a.email AS account_email,
    a.password_hash AS account_password_hash,
    a.created_at AS account_created_at
FROM sessions s
JOIN accounts a ON a.id = s.account_id
WHERE s.id = $1
`

type GetSessionWithAccountRow struct {
	SessionID           string             `json:"session_id"`
	SessionExpiresAt    pgtype.Timestamptz `json:"session_expires_at"`
	AccountID           int64              `json:"account_id"`
	AccountEmail        pgtype.Text        `json:"account_email"`
	AccountPasswordHash pgtype.Text        `json:"account_password_hash"`
	AccountCreatedAt    pgtype.Timestamptz `json:"account_created_at"`
}

func (q *Queries) GetSessionWithAccount(ctx context.Context, id string) (GetSessionWithAccountRow, error) {
	row := q.db.QueryRow(ctx, getSessionWithAccount, id)
	var i GetSessionWithAccountRow
	err := row.Scan(
		&i.SessionID,
		&i.SessionExpiresAt,
		&i.AccountID,
		&i.AccountEmail,
		&i.AccountPasswordHash,
		&i.AccountCreatedAt,
	)
	return i, err
}

const getTelegramLink = `-- name: GetTelegramLink :one
SELECT telegram_id, account_id, created_at FROM telegram_links WHERE telegram_id = $1
`

func (q *Queries) GetTelegramLink(ctx context.Context, telegramID int64) (TelegramLink, error) {
	row := q.db.QueryRow(ctx, getTelegramLink, telegramID)
	var i TelegramLink
	err := row.Scan(&i.TelegramID, &i.AccountID, &i.CreatedAt)
	return i, err
}

const lockTelegramID = `-- name: LockTelegramID :exec
SELECT pg_advisory_xact_lock($1::bigint)
`

func (q *Queries) LockTelegramID(ctx context.Context, telegramID int64) error {
	_, err := q.db.Exec(ctx, lockTelegramID, telegramID)
	return err
}

const markLinkCodeUsed = `-- name: MarkLinkCodeUsed :exec
UPDATE link_codes SET used_at = NOW() WHERE code = $1
`

func (q *Queries) MarkLinkCodeUsed(ctx context.Context, code string) error {
	_, err := q.db.Exec(ctx, markLinkCodeUsed, code)
	return err
}

const migrateSessionsToAccount = `-- name: MigrateSessionsToAccount :exec
UPDATE sessions SET account_id = $1 WHERE account_id = $2
`

type MigrateSessionsToAccountParams struct {
	ToAccountID   int64 `json:"to_account_id"`
	FromAccountID int64 `json:"from_account_id"`
}

func (q *Queries) MigrateSessionsToAccount(ctx context.Context, arg MigrateSessionsToAccountParams) error {
	_, err := q.db.Exec(ctx, migrateSessionsToAccount, arg.ToAccountID, arg.FromAccountID)
	return err
}

const updateAccountCredentials = `-- name: UpdateAccountCredentials :exec
UPDATE accounts SET email = $2, password_hash = $3 WHERE id = $1
`

type UpdateAccountCredentialsParams struct {
	ID           int64       `json:"id"`
	Email        pgtype.Text `json:"email"`
	PasswordHash pgtype.Text `json:"password_hash"`
}

func (q *Queries) UpdateAccountCredentials(ctx context.Context, arg UpdateAccountCredentialsParams) error {
	_, err := q.db.Exec(ctx, updateAccountCredentials, arg.ID, arg.Email, arg.PasswordHash)
	return err
}
