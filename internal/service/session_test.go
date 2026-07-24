package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestSessionService_Create_StoresSessionAndReturnsToken(t *testing.T) {
	var stored db.CreateSessionParams
	fq := &fakeQuerier{
		createSessionFn: func(ctx context.Context, arg db.CreateSessionParams) error {
			stored = arg
			return nil
		},
	}
	svc := service.NewSessionService(fq, time.Hour)

	token, expiresAt, err := svc.Create(context.Background(), 42)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token == "" {
		t.Fatal("expected a non-empty token")
	}
	if stored.ID != token {
		t.Errorf("expected the stored session ID to match the returned token")
	}
	if stored.AccountID != 42 {
		t.Errorf("expected account ID 42, got %d", stored.AccountID)
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expected expiresAt to be in the future")
	}
}

func TestSessionService_Create_TokensAreNotReused(t *testing.T) {
	fq := &fakeQuerier{}
	svc := service.NewSessionService(fq, time.Hour)

	token1, _, err := svc.Create(context.Background(), 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	token2, _, err := svc.Create(context.Background(), 1)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if token1 == token2 {
		t.Error("expected two calls to Create to generate different tokens")
	}
}

func TestSessionService_Validate_Success(t *testing.T) {
	fq := &fakeQuerier{
		getSessionWithAccountFn: func(ctx context.Context, id string) (db.GetSessionWithAccountRow, error) {
			return db.GetSessionWithAccountRow{
				SessionID:        id,
				SessionExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
				AccountID:        7,
				AccountEmail:     pgtype.Text{String: "a@example.com", Valid: true},
			}, nil
		},
	}
	svc := service.NewSessionService(fq, time.Hour)

	account, err := svc.Validate(context.Background(), "some-token")
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if account.ID != 7 {
		t.Errorf("expected account ID 7, got %d", account.ID)
	}
}

func TestSessionService_Validate_NotFound(t *testing.T) {
	fq := &fakeQuerier{
		getSessionWithAccountFn: func(ctx context.Context, id string) (db.GetSessionWithAccountRow, error) {
			return db.GetSessionWithAccountRow{}, pgx.ErrNoRows
		},
	}
	svc := service.NewSessionService(fq, time.Hour)

	_, err := svc.Validate(context.Background(), "nonexistent")
	if err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSessionService_Validate_ExpiredSessionIsTreatedAsNotFound(t *testing.T) {
	fq := &fakeQuerier{
		getSessionWithAccountFn: func(ctx context.Context, id string) (db.GetSessionWithAccountRow, error) {
			return db.GetSessionWithAccountRow{
				SessionID:        id,
				SessionExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
				AccountID:        7,
			}, nil
		},
	}
	svc := service.NewSessionService(fq, time.Hour)

	_, err := svc.Validate(context.Background(), "expired-token")
	if err != service.ErrNotFound {
		t.Fatalf("expected ErrNotFound for an expired session, got %v", err)
	}
}

func TestSessionService_Delete_SucceedsEvenIfAlreadyGone(t *testing.T) {
	fq := &fakeQuerier{
		deleteSessionFn: func(ctx context.Context, id string) error {
			return nil
		},
	}
	svc := service.NewSessionService(fq, time.Hour)

	if err := svc.Delete(context.Background(), "whatever"); err != nil {
		t.Fatalf("expected Delete to succeed regardless, got %v", err)
	}
}
