package service_test

import (
	"context"
	"errors"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeQuerier struct {
	countTelegramLinksForAccountFn func(ctx context.Context, accountID int64) (int64, error)
	createAccountFn                func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error)
	createEmptyAccountFn           func(ctx context.Context) (db.Account, error)
	createLinkCodeFn               func(ctx context.Context, arg db.CreateLinkCodeParams) error
	createSessionFn                func(ctx context.Context, arg db.CreateSessionParams) error
	createTelegramLinkFn           func(ctx context.Context, arg db.CreateTelegramLinkParams) error
	deleteAccountFn                func(ctx context.Context, id int64) error
	deleteExpiredSessionsFn        func(ctx context.Context) (int64, error)
	deleteSessionFn                func(ctx context.Context, id string) error
	getAccountByEmailFn            func(ctx context.Context, email pgtype.Text) (db.Account, error)
	getAccountByIDFn               func(ctx context.Context, id int64) (db.Account, error)
	getLinkCodeFn                  func(ctx context.Context, code string) (db.LinkCode, error)
	getSessionWithAccountFn        func(ctx context.Context, id string) (db.GetSessionWithAccountRow, error)
	getTelegramLinkFn              func(ctx context.Context, telegramID int64) (db.TelegramLink, error)
	lockTelegramIDFn               func(ctx context.Context, telegramID int64) error
	markLinkCodeUsedFn             func(ctx context.Context, code string) error
	migrateSessionsToAccountFn     func(ctx context.Context, arg db.MigrateSessionsToAccountParams) error
	updateAccountCredentialsFn     func(ctx context.Context, arg db.UpdateAccountCredentialsParams) error
}

var _ db.Querier = (*fakeQuerier)(nil)

func (f *fakeQuerier) CountTelegramLinksForAccount(ctx context.Context, accountID int64) (int64, error) {
	if f.countTelegramLinksForAccountFn != nil {
		return f.countTelegramLinksForAccountFn(ctx, accountID)
	}
	return 0, nil
}

func (f *fakeQuerier) CreateAccount(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
	if f.createAccountFn != nil {
		return f.createAccountFn(ctx, arg)
	}
	return db.Account{}, nil
}

func (f *fakeQuerier) CreateEmptyAccount(ctx context.Context) (db.Account, error) {
	if f.createEmptyAccountFn != nil {
		return f.createEmptyAccountFn(ctx)
	}
	return db.Account{}, nil
}

func (f *fakeQuerier) CreateLinkCode(ctx context.Context, arg db.CreateLinkCodeParams) error {
	if f.createLinkCodeFn != nil {
		return f.createLinkCodeFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) CreateSession(ctx context.Context, arg db.CreateSessionParams) error {
	if f.createSessionFn != nil {
		return f.createSessionFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) CreateTelegramLink(ctx context.Context, arg db.CreateTelegramLinkParams) error {
	if f.createTelegramLinkFn != nil {
		return f.createTelegramLinkFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) DeleteAccount(ctx context.Context, id int64) error {
	if f.deleteAccountFn != nil {
		return f.deleteAccountFn(ctx, id)
	}
	return nil
}

func (f *fakeQuerier) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	if f.deleteExpiredSessionsFn != nil {
		return f.deleteExpiredSessionsFn(ctx)
	}
	return 0, nil
}

func (f *fakeQuerier) DeleteSession(ctx context.Context, id string) error {
	if f.deleteSessionFn != nil {
		return f.deleteSessionFn(ctx, id)
	}
	return nil
}

func (f *fakeQuerier) GetAccountByEmail(ctx context.Context, email pgtype.Text) (db.Account, error) {
	if f.getAccountByEmailFn != nil {
		return f.getAccountByEmailFn(ctx, email)
	}
	return db.Account{}, errors.New("not stubbed")
}

func (f *fakeQuerier) GetAccountByID(ctx context.Context, id int64) (db.Account, error) {
	if f.getAccountByIDFn != nil {
		return f.getAccountByIDFn(ctx, id)
	}
	return db.Account{}, errors.New("not stubbed")
}

func (f *fakeQuerier) GetLinkCode(ctx context.Context, code string) (db.LinkCode, error) {
	if f.getLinkCodeFn != nil {
		return f.getLinkCodeFn(ctx, code)
	}
	return db.LinkCode{}, errors.New("not stubbed")
}

func (f *fakeQuerier) GetSessionWithAccount(ctx context.Context, id string) (db.GetSessionWithAccountRow, error) {
	if f.getSessionWithAccountFn != nil {
		return f.getSessionWithAccountFn(ctx, id)
	}
	return db.GetSessionWithAccountRow{}, errors.New("not stubbed")
}

func (f *fakeQuerier) GetTelegramLink(ctx context.Context, telegramID int64) (db.TelegramLink, error) {
	if f.getTelegramLinkFn != nil {
		return f.getTelegramLinkFn(ctx, telegramID)
	}
	return db.TelegramLink{}, errors.New("not stubbed")
}

func (f *fakeQuerier) LockTelegramID(ctx context.Context, telegramID int64) error {
	if f.lockTelegramIDFn != nil {
		return f.lockTelegramIDFn(ctx, telegramID)
	}
	return nil
}

func (f *fakeQuerier) MarkLinkCodeUsed(ctx context.Context, code string) error {
	if f.markLinkCodeUsedFn != nil {
		return f.markLinkCodeUsedFn(ctx, code)
	}
	return nil
}

func (f *fakeQuerier) MigrateSessionsToAccount(ctx context.Context, arg db.MigrateSessionsToAccountParams) error {
	if f.migrateSessionsToAccountFn != nil {
		return f.migrateSessionsToAccountFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) UpdateAccountCredentials(ctx context.Context, arg db.UpdateAccountCredentialsParams) error {
	if f.updateAccountCredentialsFn != nil {
		return f.updateAccountCredentialsFn(ctx, arg)
	}
	return nil
}
