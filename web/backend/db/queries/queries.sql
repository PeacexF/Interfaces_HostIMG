-- name: CreateAccount :one
INSERT INTO accounts (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: CreateEmptyAccount :one
-- Used for Telegram-first auto-provisioning: no email/password yet.
INSERT INTO accounts DEFAULT VALUES
RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = $1;

-- name: GetAccountByEmail :one
SELECT * FROM accounts WHERE email = $1;

-- name: UpdateAccountCredentials :exec
UPDATE accounts SET email = $2, password_hash = $3 WHERE id = $1;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1;

-- name: CreateSession :exec
INSERT INTO sessions (id, account_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetSessionWithAccount :one
SELECT
    s.id AS session_id,
    s.expires_at AS session_expires_at,
    a.id AS account_id,
    a.email AS account_email,
    a.password_hash AS account_password_hash,
    a.created_at AS account_created_at
FROM sessions s
JOIN accounts a ON a.id = s.account_id
WHERE s.id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: MigrateSessionsToAccount :exec
UPDATE sessions SET account_id = sqlc.arg(to_account_id) WHERE account_id = sqlc.arg(from_account_id);

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at < NOW();

-- name: GetTelegramLink :one
SELECT * FROM telegram_links WHERE telegram_id = $1;

-- name: CreateTelegramLink :exec
INSERT INTO telegram_links (telegram_id, account_id)
VALUES ($1, $2);

-- name: CountTelegramLinksForAccount :one
SELECT COUNT(*) FROM telegram_links WHERE account_id = $1;

-- name: LockTelegramID :exec
SELECT pg_advisory_xact_lock(sqlc.arg(telegram_id)::bigint);

-- name: CreateLinkCode :exec
INSERT INTO link_codes (code, account_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetLinkCode :one
SELECT * FROM link_codes WHERE code = $1;

-- name: MarkLinkCodeUsed :exec
UPDATE link_codes SET used_at = NOW() WHERE code = $1;
