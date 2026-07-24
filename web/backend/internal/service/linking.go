package service

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

const linkCodeAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"
const linkCodeLength = 8

type LinkingService struct {
	dbPool  *pgxpool.Pool
	queries TxQuerier
	codeTTL time.Duration
}

func NewLinkingService(dbPool *pgxpool.Pool, queries TxQuerier, codeTTL time.Duration) *LinkingService {
	return &LinkingService{dbPool: dbPool, queries: queries, codeTTL: codeTTL}
}

func (l *LinkingService) StartLink(ctx context.Context, accountID int64) (code string, expiresAt time.Time, err error) {
	code, err = generateLinkCode()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt = time.Now().Add(l.codeTTL)

	if err := l.queries.CreateLinkCode(ctx, db.CreateLinkCodeParams{
		Code:      code,
		AccountID: accountID,
		ExpiresAt: pgtypeTimestamptz(expiresAt),
	}); err != nil {
		return "", time.Time{}, err
	}
	return code, expiresAt, nil
}

func (l *LinkingService) CompleteLink(ctx context.Context, code string, telegramID int64) (survivingAccountID int64, err error) {
	tx, err := l.dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	qtx := l.queries.WithTx(tx)

	linkCode, err := qtx.GetLinkCode(ctx, code)
	if err != nil {
		if isNoRows(err) {
			return 0, ErrLinkCodeInvalid
		}
		return 0, err
	}
	if linkCode.UsedAt.Valid {
		return 0, ErrLinkCodeInvalid
	}
	if linkCode.ExpiresAt.Valid && linkCode.ExpiresAt.Time.Before(time.Now()) {
		return 0, ErrLinkCodeInvalid
	}

	requestingID := linkCode.AccountID

	existingLink, err := qtx.GetTelegramLink(ctx, telegramID)
	switch {
	case isNoRows(err):
		if err := qtx.CreateTelegramLink(ctx, db.CreateTelegramLinkParams{
			TelegramID: telegramID,
			AccountID:  requestingID,
		}); err != nil {
			return 0, err
		}
		if err := qtx.MarkLinkCodeUsed(ctx, code); err != nil {
			return 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return requestingID, nil

	case err != nil:
		return 0, err
	}

	survivingID := existingLink.AccountID
	if survivingID == requestingID {
		if err := qtx.MarkLinkCodeUsed(ctx, code); err != nil {
			return 0, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, err
		}
		return survivingID, nil
	}

	survivingAccount, err := qtx.GetAccountByID(ctx, survivingID)
	if err != nil {
		return 0, err
	}
	requestingAccount, err := qtx.GetAccountByID(ctx, requestingID)
	if err != nil {
		return 0, err
	}
	requestingOtherLinks, err := qtx.CountTelegramLinksForAccount(ctx, requestingID)
	if err != nil {
		return 0, err
	}

	if survivingAccount.Email.Valid || requestingOtherLinks > 0 {
		return 0, ErrAccountsBothHaveHistory
	}

	if err := qtx.MigrateSessionsToAccount(ctx, db.MigrateSessionsToAccountParams{
		FromAccountID: requestingID,
		ToAccountID:   survivingID,
	}); err != nil {
		return 0, err
	}
	if err := qtx.DeleteAccount(ctx, requestingID); err != nil {
		return 0, err
	}
	if err := qtx.UpdateAccountCredentials(ctx, db.UpdateAccountCredentialsParams{
		ID:           survivingID,
		Email:        requestingAccount.Email,
		PasswordHash: requestingAccount.PasswordHash,
	}); err != nil {
		return 0, err
	}
	if err := qtx.MarkLinkCodeUsed(ctx, code); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return survivingID, nil
}

func generateLinkCode() (string, error) {
	buf := make([]byte, linkCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := make([]byte, linkCodeLength)
	for i, b := range buf {
		code[i] = linkCodeAlphabet[int(b)%len(linkCodeAlphabet)]
	}
	return string(code), nil
}
