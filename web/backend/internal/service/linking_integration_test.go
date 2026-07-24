//go:build integration

package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestLinkingService_CompleteLink_SimpleCase(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	linking := service.NewLinkingService(pool, queries, time.Hour)
	accounts := service.NewAccountService(queries, service.NewSessionService(queries, time.Hour))

	requestingID, _, _, err := accounts.Signup(context.Background(), "fresh@example.com", "password123")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	code, _, err := linking.StartLink(context.Background(), requestingID)
	if err != nil {
		t.Fatalf("StartLink: %v", err)
	}

	const telegramID = int64(111111111)
	survivingID, err := linking.CompleteLink(context.Background(), code, telegramID)
	if err != nil {
		t.Fatalf("CompleteLink: %v", err)
	}
	if survivingID != requestingID {
		t.Errorf("expected the requesting account to survive in the simple case, got %d want %d", survivingID, requestingID)
	}

	link, err := queries.GetTelegramLink(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("GetTelegramLink: %v", err)
	}
	if link.AccountID != requestingID {
		t.Errorf("expected telegram_links to point at %d, got %d", requestingID, link.AccountID)
	}
}

func TestLinkingService_CompleteLink_MergeCase(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	identity := service.NewIdentityService(pool, queries)
	sessions := service.NewSessionService(queries, time.Hour)
	accounts := service.NewAccountService(queries, sessions)
	linking := service.NewLinkingService(pool, queries, time.Hour)

	const telegramID = int64(222222222)

	preExistingID, err := identity.ResolveTelegramUser(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("ResolveTelegramUser: %v", err)
	}

	requestingID, sessionToken, _, err := accounts.Signup(context.Background(), "newsignup@example.com", "password123")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if requestingID == preExistingID {
		t.Fatalf("test setup bug: requesting and pre-existing accounts must differ")
	}

	code, _, err := linking.StartLink(context.Background(), requestingID)
	if err != nil {
		t.Fatalf("StartLink: %v", err)
	}

	survivingID, err := linking.CompleteLink(context.Background(), code, telegramID)
	if err != nil {
		t.Fatalf("CompleteLink: %v", err)
	}
	if survivingID != preExistingID {
		t.Errorf("expected the pre-existing Telegram account to survive the merge, got %d want %d", survivingID, preExistingID)
	}

	if _, err := queries.GetAccountByID(context.Background(), requestingID); err == nil {
		t.Error("expected the discarded requesting account to no longer exist")
	}

	survivingAccount, err := queries.GetAccountByID(context.Background(), preExistingID)
	if err != nil {
		t.Fatalf("GetAccountByID: %v", err)
	}
	if survivingAccount.Email.String != "newsignup@example.com" {
		t.Errorf("expected surviving account to have the signup's email, got %q", survivingAccount.Email.String)
	}

	migratedAccount, err := sessions.Validate(context.Background(), sessionToken)
	if err != nil {
		t.Fatalf("expected the pre-link session to still validate after merge, got error: %v", err)
	}
	if migratedAccount.ID != preExistingID {
		t.Errorf("expected the migrated session to now point at %d, got %d", preExistingID, migratedAccount.ID)
	}

	if _, _, err := accounts.Login(context.Background(), "newsignup@example.com", "password123"); err != nil {
		t.Errorf("expected login with the moved credentials to succeed, got %v", err)
	}
}

func TestLinkingService_CompleteLink_RefusesWhenBothSidesHaveHistory(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	identity := service.NewIdentityService(pool, queries)
	accounts := service.NewAccountService(queries, service.NewSessionService(queries, time.Hour))
	linking := service.NewLinkingService(pool, queries, time.Hour)

	const telegramID = int64(333333333)

	preExistingID, err := identity.ResolveTelegramUser(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("ResolveTelegramUser: %v", err)
	}
	if err := queries.UpdateAccountCredentials(context.Background(), db.UpdateAccountCredentialsParams{
		ID:           preExistingID,
		Email:        pgtypeText("already-has-login@example.com"),
		PasswordHash: pgtypeText("somehash"),
	}); err != nil {
		t.Fatalf("simulate pre-existing credentials: %v", err)
	}

	requestingID, _, _, err := accounts.Signup(context.Background(), "other@example.com", "password123")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	code, _, err := linking.StartLink(context.Background(), requestingID)
	if err != nil {
		t.Fatalf("StartLink: %v", err)
	}

	_, err = linking.CompleteLink(context.Background(), code, telegramID)
	if err != service.ErrAccountsBothHaveHistory {
		t.Fatalf("expected ErrAccountsBothHaveHistory, got %v", err)
	}

	if _, err := queries.GetAccountByID(context.Background(), requestingID); err != nil {
		t.Errorf("expected the requesting account to still exist after a refused merge, got %v", err)
	}
}

func TestLinkingService_CompleteLink_InvalidCode(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	linking := service.NewLinkingService(pool, queries, time.Hour)

	_, err := linking.CompleteLink(context.Background(), "does-not-exist", 444444444)
	if err != service.ErrLinkCodeInvalid {
		t.Fatalf("expected ErrLinkCodeInvalid, got %v", err)
	}
}

func TestLinkingService_CompleteLink_ExpiredCode(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	accounts := service.NewAccountService(queries, service.NewSessionService(queries, time.Hour))
	linking := service.NewLinkingService(pool, queries, -time.Minute)

	requestingID, _, _, err := accounts.Signup(context.Background(), "expiretest@example.com", "password123")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	code, _, err := linking.StartLink(context.Background(), requestingID)
	if err != nil {
		t.Fatalf("StartLink: %v", err)
	}

	_, err = linking.CompleteLink(context.Background(), code, 555555555)
	if err != service.ErrLinkCodeInvalid {
		t.Fatalf("expected ErrLinkCodeInvalid for an expired code, got %v", err)
	}
}

func pgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}
