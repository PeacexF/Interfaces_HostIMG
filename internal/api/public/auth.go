package public

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
)

type AuthHandler struct {
	accounts     *service.AccountService
	sessions     *service.SessionService
	cookieSecure bool
}

func NewAuthHandler(accounts *service.AccountService, sessions *service.SessionService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{accounts: accounts, sessions: sessions, cookieSecure: cookieSecure}
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const minPasswordLength = 8

func (h *AuthHandler) HandleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	if len(req.Password) < minPasswordLength {
		writeJSONError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	accountID, token, expiresAt, err := h.accounts.Signup(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			writeJSONError(w, http.StatusConflict, "email is already in use")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	setSessionCookie(w, token, expiresAt, h.cookieSecure)
	writeJSON(w, http.StatusCreated, map[string]any{"account_id": accountID})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, expiresAt, err := h.accounts.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeJSONError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	setSessionCookie(w, token, expiresAt, h.cookieSecure)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil && cookie.Value != "" {

		_ = h.sessions.Delete(r.Context(), cookie.Value)
	}
	clearSessionCookie(w, h.cookieSecure)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *AuthHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	accountID, ok := middleware.AccountIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not logged in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account_id": accountID})
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
