//go:build integration

package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBPool(t *testing.T) (*pgxpool.Pool, *db.Queries) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool, db.New(pool)
}

func cleanupTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "TRUNCATE accounts, sessions, telegram_links, link_codes RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

func TestIdentityService_ResolveTelegramUser_Integration(t *testing.T) {
	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	identity := service.NewIdentityService(pool, queries)

	const telegramID = int64(123456789)

	accountID, err := identity.ResolveTelegramUser(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("ResolveTelegramUser (first call): %v", err)
	}
	if accountID == 0 {
		t.Fatal("expected a non-zero account ID")
	}

	accountID2, err := identity.ResolveTelegramUser(context.Background(), telegramID)
	if err != nil {
		t.Fatalf("ResolveTelegramUser (second call): %v", err)
	}
	if accountID2 != accountID {
		t.Errorf("expected the same account ID on re-resolve, got %d then %d", accountID, accountID2)
	}
}

func TestIdentityService_ResolveTelegramUser_ConcurrentFirstContact(t *testing.T) {

	pool, queries := testDBPool(t)
	cleanupTables(t, pool)

	identity := service.NewIdentityService(pool, queries)

	const telegramID = int64(987654321)

	results := make(chan int64, 2)
	errs := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			id, err := identity.ResolveTelegramUser(context.Background(), telegramID)
			if err != nil {
				errs <- err
				return
			}
			results <- id
		}()
	}

	var ids []int64
	for i := 0; i < 2; i++ {
		select {
		case err := <-errs:
			t.Fatalf("concurrent ResolveTelegramUser failed: %v", err)
		case id := <-results:
			ids = append(ids, id)
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for concurrent resolves")
		}
	}

	if ids[0] != ids[1] {
		t.Errorf("expected both concurrent calls to resolve to the same account, got %d and %d", ids[0], ids[1])
	}
}
