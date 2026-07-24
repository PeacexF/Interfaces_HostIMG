package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
)

type contextKey string

const AccountIDKey contextKey = "account_id"

const SessionCookieName = "session"

type SessionValidator interface {
	Validate(ctx context.Context, token string) (accountID int64, err error)
}

type sessionServiceAdapter struct {
	svc *service.SessionService
}

func NewSessionValidator(svc *service.SessionService) SessionValidator {
	return sessionServiceAdapter{svc: svc}
}

func (a sessionServiceAdapter) Validate(ctx context.Context, token string) (int64, error) {
	account, err := a.svc.Validate(ctx, token)
	if err != nil {
		return 0, err
	}
	return account.ID, nil
}

func RequireSession(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				writeJSONError(w, http.StatusUnauthorized, "not logged in")
				return
			}

			accountID, err := validator.Validate(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, service.ErrNotFound) {
					writeJSONError(w, http.StatusUnauthorized, "session expired or invalid")
					return
				}
				writeJSONError(w, http.StatusInternalServerError, "failed to validate session")
				return
			}

			ctx := context.WithValue(r.Context(), AccountIDKey, accountID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccountIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(AccountIDKey).(int64)
	return id, ok
}
