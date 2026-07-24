package public_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PeacexF/Interfaces_HostIMG/internal/api/http/middleware"
	"github.com/PeacexF/Interfaces_HostIMG/internal/api/public"
	"github.com/PeacexF/Interfaces_HostIMG/internal/service"
	"github.com/go-chi/chi/v5"
)

func withAccount(r *http.Request, id int64) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.AccountIDKey, id)
	return r.WithContext(ctx)
}

func TestFilesHandler_HandleDownload_ForwardsUserID(t *testing.T) {
	var gotUserID, gotToken, gotPath string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get("X-User-ID")
		gotToken = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("bytes"))
	}))
	defer fake.Close()

	client := service.NewHostIMGClient(fake.URL, "test-token")
	handler := public.NewFilesHandler(client)

	r := chi.NewRouter()
	r.Get("/files/{id}", handler.HandleDownload)

	req := withAccount(httptest.NewRequest(http.MethodGet, "/files/abc123", nil), 42)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "bytes" {
		t.Errorf("expected body to be forwarded, got %q", rec.Body.String())
	}
	if gotUserID != "42" {
		t.Errorf("expected X-User-ID 42, got %q", gotUserID)
	}
	if gotToken != "Bearer test-token" {
		t.Errorf("expected Authorization forwarded, got %q", gotToken)
	}
	if gotPath != "/api/v1/objects/abc123" {
		t.Errorf("expected path /api/v1/objects/abc123, got %q", gotPath)
	}
}

func TestFilesHandler_RequiresSession(t *testing.T) {
	client := service.NewHostIMGClient("http://unused", "token")
	handler := public.NewFilesHandler(client)

	r := chi.NewRouter()
	r.Get("/files/{id}", handler.HandleDownload)

	req := httptest.NewRequest(http.MethodGet, "/files/abc123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestFilesHandler_HandleList_ForwardsQueryParams(t *testing.T) {
	var gotQuery string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer fake.Close()

	client := service.NewHostIMGClient(fake.URL, "token")
	handler := public.NewFilesHandler(client)

	r := chi.NewRouter()
	r.Get("/files", handler.HandleList)

	req := withAccount(httptest.NewRequest(http.MethodGet, "/files?limit=10&offset=5", nil), 1)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if gotQuery != "limit=10&offset=5" {
		t.Errorf("expected query forwarded, got %q", gotQuery)
	}
}
