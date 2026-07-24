package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/internalapi"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/public"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type ServerConfig struct {
	Port                 int
	InternalSharedSecret string

	Auth     *public.AuthHandler
	Link     *public.LinkHandler
	Identity *internalapi.IdentityHandler

	SessionValidator middleware.SessionValidator
}

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg ServerConfig) *Server {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/signup", cfg.Auth.HandleSignup)
		r.Post("/login", cfg.Auth.HandleLogin)
		r.Post("/logout", cfg.Auth.HandleLogout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireSession(cfg.SessionValidator))
			r.Get("/me", cfg.Auth.HandleMe)
			r.Post("/link/start", cfg.Link.HandleStartLink)
		})
	})

	r.Route("/internal", func(r chi.Router) {
		r.Use(middleware.RequireInternalSecret(cfg.InternalSharedSecret))
		r.Post("/resolve-telegram-user", cfg.Identity.HandleResolve)
		r.Post("/complete-link", cfg.Identity.HandleCompleteLink)
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: r,
		},
	}
}

func (s *Server) Start() error {
	slog.Info("website backend listening", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}
