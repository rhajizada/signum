package router_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/requestctx"
	"github.com/rhajizada/signum/internal/router"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/pkg/renderer"
	"golang.org/x/image/font/basicfont"
)

func TestHandleSetsRoutePattern(t *testing.T) {
	r := router.New(newHandler(t))

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
	r := router.New(newHandler(t))
	var calls []string

	wrap := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, req)
			})
		}
	}

	r.Handle("GET /order", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
	if ctx == nil {
		return repository.Badge{}, errors.New("missing context")
	}
	if arg.Subject == "" || arg.Status == "" || arg.Color == "" || arg.Style == "" {
		return repository.Badge{}, errors.New("missing badge fields")
	}
	return repository.Badge{}, nil
}

func (f *fakeRepo) GetBadgeByID(ctx context.Context, id uuid.UUID) (repository.Badge, error) {
	if ctx == nil {
		return repository.Badge{}, errors.New("missing context")
	}
	if id == uuid.Nil {
		return repository.Badge{}, errors.New("missing id")
	}
	return repository.Badge{}, sql.ErrNoRows
}

func (f *fakeRepo) UpdateBadge(ctx context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error) {
	if ctx == nil {
		return repository.Badge{}, errors.New("missing context")
	}
	if arg.ID == uuid.Nil {
		return repository.Badge{}, errors.New("missing id")
	}
	return repository.Badge{}, nil
}

func (f *fakeRepo) DeleteBadge(ctx context.Context, id uuid.UUID) error {
	if ctx == nil {
		return errors.New("missing context")
	}
	if id == uuid.Nil {
		return errors.New("missing id")
	}
	return nil
}

func newHandler(tb testing.TB) *handler.Handler {
	tb.Helper()
	rdr, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	if err != nil {
		tb.Fatalf("renderer: %v", err)
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		tb.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(rdr, &fakeRepo{}, tokens)
	if err != nil {
		tb.Fatalf("service: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := handler.New(svc, logger)
	if err != nil {
		tb.Fatalf("handler: %v", err)
	}
	return h
}

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

	r := router.New(h)
	req := httptest.NewRequest(http.MethodGet, "/api/badges/live?subject=build&status=passing&color=green", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") == "" {
		t.Fatalf("expected content type to be set")
	}
}
