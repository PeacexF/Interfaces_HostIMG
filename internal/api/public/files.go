package public

import (
	"io"
	"net/http"

	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/go-chi/chi/v5"
)

type FilesHandler struct {
	hostimg *service.HostIMGClient
}

func NewFilesHandler(hostimg *service.HostIMGClient) *FilesHandler {
	return &FilesHandler{hostimg: hostimg}
}

func (h *FilesHandler) proxy(w http.ResponseWriter, r *http.Request, method, path string) {
	accountID, ok := middleware.AccountIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not logged in")
		return
	}

	resp, err := h.hostimg.Do(r.Context(), method, path, accountID, r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "hostimg request failed")
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (h *FilesHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodPost, "/api/v1/objects")
}

func (h *FilesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	h.proxy(w, r, http.MethodGet, "/api/v1/objects?"+r.URL.RawQuery)
}

func (h *FilesHandler) HandleMeta(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.proxy(w, r, http.MethodGet, "/api/v1/objects/"+id+"/meta")
}

func (h *FilesHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.proxy(w, r, http.MethodGet, "/api/v1/objects/"+id)
}

func (h *FilesHandler) HandleThumbnail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.proxy(w, r, http.MethodGet, "/api/v1/objects/"+id+"/thumbnail?"+r.URL.RawQuery)
}

func (h *FilesHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.proxy(w, r, http.MethodDelete, "/api/v1/objects/"+id)
}
