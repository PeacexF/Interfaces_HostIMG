package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAccountService_Signup_Success(t *testing.T) {
	var createdParams db.CreateAccountParams
	fq := &fakeQuerier{
		getAccountByEmailFn: func(ctx context.Context, email pgtype.Text) (db.Account, error) {
			return db.Account{}, pgx.ErrNoRows
		},
		createAccountFn: func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
			createdParams = arg
			return db.Account{ID: 1, Email: arg.Email, PasswordHash: arg.PasswordHash}, nil
		},
	}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	accountID, token, expiresAt, err := accounts.Signup(context.Background(), "new@example.com", "hunter22")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if accountID != 1 {
		t.Errorf("expected account ID 1, got %d", accountID)
	}
	if token == "" {
		t.Error("expected a non-empty session token")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expected expiresAt in the future")
	}
	if createdParams.Email.String != "new@example.com" {
		t.Errorf("expected email to be passed through, got %q", createdParams.Email.String)
	}
	if createdParams.PasswordHash.String == "hunter22" {
		t.Error("expected the password to be hashed, not stored in plain text")
	}
}

func TestAccountService_Signup_RejectsTakenEmail(t *testing.T) {
	fq := &fakeQuerier{
		getAccountByEmailFn: func(ctx context.Context, email pgtype.Text) (db.Account, error) {
			return db.Account{ID: 5}, nil
		},
	}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	_, _, _, err := accounts.Signup(context.Background(), "taken@example.com", "hunter22")
	if err != service.ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestAccountService_Signup_RaceOnEmailUniquenessIsTranslatedToErrEmailTaken(t *testing.T) {

	fq := &fakeQuerier{
		getAccountByEmailFn: func(ctx context.Context, email pgtype.Text) (db.Account, error) {
			return db.Account{}, pgx.ErrNoRows
		},
		createAccountFn: func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
			return db.Account{}, &pgconn.PgError{Code: "23505"}
		},
	}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	_, _, _, err := accounts.Signup(context.Background(), "racy@example.com", "hunter22")
	if err != service.ErrEmailTaken {
		t.Fatalf("expected the race to be translated to ErrEmailTaken, got %v", err)
	}
}

func TestAccountService_Login_Success(t *testing.T) {
	fq := &fakeQuerier{}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	fq.getAccountByEmailFn = func(ctx context.Context, email pgtype.Text) (db.Account, error) {
		return db.Account{}, pgx.ErrNoRows
	}
	var createdAccount db.Account
	fq.createAccountFn = func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
		createdAccount = db.Account{ID: 1, Email: arg.Email, PasswordHash: arg.PasswordHash}
		return createdAccount, nil
	}
	if _, _, _, err := accounts.Signup(context.Background(), "login@example.com", "correct-password"); err != nil {
		t.Fatalf("signup fixture setup failed: %v", err)
	}

	fq.getAccountByEmailFn = func(ctx context.Context, email pgtype.Text) (db.Account, error) {
		return createdAccount, nil
	}

	token, expiresAt, err := accounts.Login(context.Background(), "login@example.com", "correct-password")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Error("expected a non-empty session token")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expected expiresAt in the future")
	}
}

func TestAccountService_Login_WrongPassword(t *testing.T) {
	fq := &fakeQuerier{}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	var createdAccount db.Account
	fq.getAccountByEmailFn = func(ctx context.Context, email pgtype.Text) (db.Account, error) {
		return db.Account{}, pgx.ErrNoRows
	}
	fq.createAccountFn = func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
		createdAccount = db.Account{ID: 1, Email: arg.Email, PasswordHash: arg.PasswordHash}
		return createdAccount, nil
	}
	if _, _, _, err := accounts.Signup(context.Background(), "user@example.com", "correct-password"); err != nil {
		t.Fatalf("signup fixture setup failed: %v", err)
	}

	fq.getAccountByEmailFn = func(ctx context.Context, email pgtype.Text) (db.Account, error) {
		return createdAccount, nil
	}

	_, _, err := accounts.Login(context.Background(), "user@example.com", "wrong-password")
	if err != service.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAccountService_Login_UnknownEmail(t *testing.T) {
	fq := &fakeQuerier{
		getAccountByEmailFn: func(ctx context.Context, email pgtype.Text) (db.Account, error) {
			return db.Account{}, pgx.ErrNoRows
		},
	}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	_, _, err := accounts.Login(context.Background(), "nobody@example.com", "anything")
	if err != service.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials (not a distinguishable 'no such user' error), got %v", err)
	}
}

func TestAccountService_Login_TelegramOnlyAccountCannotLogInWithoutAPassword(t *testing.T) {

	fq := &fakeQuerier{
		getAccountByEmailFn: func(ctx context.Context, email pgtype.Text) (db.Account, error) {
			return db.Account{ID: 9, PasswordHash: pgtype.Text{Valid: false}}, nil
		},
	}
	sessions := service.NewSessionService(fq, time.Hour)
	accounts := service.NewAccountService(fq, sessions)

	_, _, err := accounts.Login(context.Background(), "telegram-only@example.com", "anything")
	if err != service.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
