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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/basicfont"
)

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
	require.NoError(tb, err)
	tokens, err := service.NewTokenManager("secret")
	require.NoError(tb, err)
	svc, err := service.New(rdr, &fakeRepo{}, tokens)
	require.NoError(tb, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := handler.New(svc, logger)
	require.NoError(tb, err)
	return h
}

func TestRouterBehaviors(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "sets route pattern on context",
			run: func(t *testing.T) {
				r := router.New(newHandler(t))
				r.Handle("GET /things/{id}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					route, ok := requestctx.RoutePattern(req.Context())
					assert.True(t, ok)
					assert.Equal(t, "GET /things/{id}", route)
					w.WriteHeader(http.StatusOK)
				}))

				req := httptest.NewRequest(http.MethodGet, "/things/123", nil)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				assert.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "applies wrappers in order",
			run: func(t *testing.T) {
				r := router.New(newHandler(t))
				calls := make([]string, 0, 3)
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

				assert.Equal(t, []string{"first", "second", "handler"}, calls)
				assert.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "wires live badge route",
			run: func(t *testing.T) {
				rdr, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
				require.NoError(t, err)
				tokens, err := service.NewTokenManager("secret")
				require.NoError(t, err)
				svc, err := service.New(rdr, &fakeRepo{}, tokens)
				require.NoError(t, err)
				logger := slog.New(slog.NewTextHandler(io.Discard, nil))
				h, err := handler.New(svc, logger)
				require.NoError(t, err)

				r := router.New(h)
				req := httptest.NewRequest(
					http.MethodGet,
					"/api/badges/live?subject=build&status=passing&color=green",
					nil,
				)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusOK, rec.Code)
				assert.NotEmpty(t, rec.Header().Get("Content-Type"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}
