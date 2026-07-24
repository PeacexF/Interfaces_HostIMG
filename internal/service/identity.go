package service

import (
	"context"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdentityService struct {
	dbPool  *pgxpool.Pool
	queries TxQuerier
}

func NewIdentityService(dbPool *pgxpool.Pool, queries TxQuerier) *IdentityService {
	return &IdentityService{dbPool: dbPool, queries: queries}
}

func (s *IdentityService) ResolveTelegramUser(ctx context.Context, telegramID int64) (accountID int64, err error) {

	existing, err := s.queries.GetTelegramLink(ctx, telegramID)
	if err == nil {
		return existing.AccountID, nil
	}
	if !isNoRows(err) {
		return 0, err
	}

	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	if err := qtx.LockTelegramID(ctx, telegramID); err != nil {
		return 0, err
	}

	existing, err = qtx.GetTelegramLink(ctx, telegramID)
	if err == nil {

		return existing.AccountID, nil
	}
	if !isNoRows(err) {
		return 0, err
	}

	account, err := qtx.CreateEmptyAccount(ctx)
	if err != nil {
		return 0, err
	}

	if err := qtx.CreateTelegramLink(ctx, db.CreateTelegramLinkParams{
		TelegramID: telegramID,
		AccountID:  account.ID,
	}); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return account.ID, nil
}
