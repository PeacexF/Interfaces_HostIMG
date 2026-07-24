package internalapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

type IdentityHandler struct {
	identity *service.IdentityService
	linking  *service.LinkingService
}

func NewIdentityHandler(identity *service.IdentityService, linking *service.LinkingService) *IdentityHandler {
	return &IdentityHandler{identity: identity, linking: linking}
}

type resolveRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

func (h *IdentityHandler) HandleResolve(w http.ResponseWriter, r *http.Request) {
	var req resolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.TelegramID == 0 {
		writeJSONError(w, http.StatusBadRequest, "telegram_id is required")
		return
	}

	accountID, err := h.identity.ResolveTelegramUser(r.Context(), req.TelegramID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to resolve identity")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"account_id": accountID})
}

type completeLinkRequest struct {
	Code       string `json:"code"`
	TelegramID int64  `json:"telegram_id"`
}

func (h *IdentityHandler) HandleCompleteLink(w http.ResponseWriter, r *http.Request) {
	var req completeLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" || req.TelegramID == 0 {
		writeJSONError(w, http.StatusBadRequest, "code and telegram_id are required")
		return
	}

	accountID, err := h.linking.CompleteLink(r.Context(), req.Code, req.TelegramID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkCodeInvalid):
			writeJSONError(w, http.StatusBadRequest, "link code is invalid, expired, or already used")
		case errors.Is(err, service.ErrAccountsBothHaveHistory):
			writeJSONError(w, http.StatusConflict, "both accounts already have independent history and cannot be automatically linked — contact support")
		default:
			writeJSONError(w, http.StatusInternalServerError, "failed to complete link")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"account_id": accountID})
}
