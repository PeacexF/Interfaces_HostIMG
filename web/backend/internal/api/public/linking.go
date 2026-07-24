package public

import (
	"net/http"

	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
)

type LinkHandler struct {
	linking *service.LinkingService
}

func NewLinkHandler(linking *service.LinkingService) *LinkHandler {
	return &LinkHandler{linking: linking}
}

func (h *LinkHandler) HandleStartLink(w http.ResponseWriter, r *http.Request) {
	accountID, ok := middleware.AccountIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	code, expiresAt, err := h.linking.StartLink(r.Context(), accountID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate link code")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":       code,
		"expires_at": expiresAt,
	})
}
