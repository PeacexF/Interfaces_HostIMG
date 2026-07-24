package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	hostimghttp "github.com/PeacexF/Interfaces_HostIMG/internal/api/http"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/internalapi"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/public"
	"github.com/PeacexF/Interfaces_HostIMG/internal/config"
	"github.com/PeacexF/Interfaces_HostIMG/internal/db"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		slog.Error("postgres ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to postgres")

	queries := db.New(dbPool)

	sessionService := service.NewSessionService(queries, cfg.SessionTTL)
	accountService := service.NewAccountService(queries, sessionService)
	identityService := service.NewIdentityService(dbPool, queries)
	linkingService := service.NewLinkingService(dbPool, queries, cfg.LinkCodeTTL)

	authHandler := public.NewAuthHandler(accountService, sessionService, cfg.SessionCookieSecure)
	linkHandler := public.NewLinkHandler(linkingService)
	identityHandler := internalapi.NewIdentityHandler(identityService, linkingService)

	srv := hostimghttp.NewServer(hostimghttp.ServerConfig{
		Port:                 cfg.Port,
		InternalSharedSecret: cfg.InternalSharedSecret,
		Auth:                 authHandler,
		Link:                 linkHandler,
		Identity:             identityHandler,
		SessionValidator:     middleware.NewSessionValidator(sessionService),
	})

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Stop(shutdownCtx); err != nil {
		slog.Error("failed to stop server cleanly", "error", err)
	}
	slog.Info("stopped")
}
