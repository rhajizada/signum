package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"io"

	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/requestctx"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/pkg/renderer"
	"golang.org/x/image/font/basicfont"
	"context"
	"github.com/google/uuid"
	"database/sql"
)

func TestHandleSetsRoutePattern(t *testing.T) {
	r := &Router{mux: http.NewServeMux()}

	r.Handle("GET /things/{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		route, ok := requestctx.RoutePattern(req.Context())
		if !ok || route != "GET /things/{id}" {
			t.Fatalf("expected route pattern to be set, got %q (ok=%v)", route, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/things/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d", rec.Code)
	}
}

func TestHandleWrapperOrder(t *testing.T) {
	r := &Router{mux: http.NewServeMux()}
	var calls []string

	wrap := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, req)
			})
		}
	}

	r.Handle("GET /order", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusOK)
	}), wrap("first"), wrap("second"))

	req := httptest.NewRequest(http.MethodGet, "/order", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	expected := []string{"first", "second", "handler"}
	if len(calls) != len(expected) {
		t.Fatalf("expected %d calls, got %d", len(expected), len(calls))
	}
	for i, name := range expected {
		if calls[i] != name {
			t.Fatalf("expected call %d to be %q, got %q", i, name, calls[i])
		}
	}
}

type fakeRepo struct{}

func (f *fakeRepo) CreateBadge(ctx context.Context, arg repository.CreateBadgeParams) (repository.Badge, error) {
	return repository.Badge{}, nil
}
func (f *fakeRepo) GetBadgeByID(ctx context.Context, id uuid.UUID) (repository.Badge, error) {
	return repository.Badge{}, sql.ErrNoRows
}
func (f *fakeRepo) UpdateBadge(ctx context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error) {
	return repository.Badge{}, nil
}
func (f *fakeRepo) DeleteBadge(ctx context.Context, id uuid.UUID) error { return nil }

func TestNewRoutesLiveBadge(t *testing.T) {
	rdr, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(rdr, &fakeRepo{}, tokens)
	if err != nil {
		t.Fatalf("service: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := handler.New(svc, logger)
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	router := New(h)
	req := httptest.NewRequest(http.MethodGet, "/api/badges/live?subject=build&status=passing&color=green", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") == "" {
		t.Fatalf("expected content type to be set")
	}
}
