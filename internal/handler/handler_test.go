package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/models"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/pkg/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/basicfont"
)

type fakeRepo struct {
	createFn func(ctx context.Context, arg repository.CreateBadgeParams) (repository.Badge, error)
	getFn    func(ctx context.Context, id uuid.UUID) (repository.Badge, error)
	updateFn func(ctx context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (f *fakeRepo) CreateBadge(ctx context.Context, arg repository.CreateBadgeParams) (repository.Badge, error) {
	if f.createFn != nil {
		return f.createFn(ctx, arg)
	}
	return repository.Badge{}, nil
}

func (f *fakeRepo) GetBadgeByID(ctx context.Context, id uuid.UUID) (repository.Badge, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return repository.Badge{}, sql.ErrNoRows
}

func (f *fakeRepo) UpdateBadge(ctx context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, arg)
	}
	return repository.Badge{}, nil
}

func (f *fakeRepo) DeleteBadge(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

func newHandler(tb testing.TB, repo service.BadgeRepository, tokens *service.TokenManager) *handler.Handler {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	require.NoError(tb, err)
	svc, err := service.New(r, repo, tokens)
	require.NoError(tb, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := handler.New(svc, logger)
	require.NoError(tb, err)
	return h
}

func newTokens(tb testing.TB) *service.TokenManager {
	tb.Helper()
	tokens, err := service.NewTokenManager("secret")
	require.NoError(tb, err)
	return tokens
}

func TestNewHandlerRequiresService(t *testing.T) {
	tests := []struct {
		name string
		svc  *service.Service
	}{
		{name: "nil service"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, err := handler.New(tc.svc, nil)
			require.Nil(t, h)
			require.EqualError(t, err, "service is required")
		})
	}
}

func TestCreateBadgeHandler(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		body       string
		repo       *fakeRepo
		wantStatus int
		assertBody func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "creates badge",
			body: `{"subject":"build","status":"passing","color":"green","style":"flat"}`,
			repo: &fakeRepo{
				createFn: func(_ context.Context, arg repository.CreateBadgeParams) (repository.Badge, error) {
					return repository.Badge{
						ID:        uuid.New(),
						TokenHash: arg.TokenHash,
						Subject:   arg.Subject,
						Status:    arg.Status,
						Color:     arg.Color,
						Style:     arg.Style,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			assertBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp models.CreateBadgeResponse
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, "build", resp.Subject)
				assert.Equal(t, "passing", resp.Status)
			},
		},
		{name: "rejects invalid json", body: `{"subject":`, repo: &fakeRepo{}, wantStatus: http.StatusBadRequest},
		{
			name:       "rejects trailing json",
			body:       `{"subject":"a"}{"extra":"b"}`,
			repo:       &fakeRepo{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "rejects oversized body",
			body: `{"subject":"build","status":"` + string(
				bytes.Repeat([]byte{'a'}, 70*1024),
			) + `","color":"green","style":"flat"}`,
			repo:       &fakeRepo{},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, tc.repo, newTokens(t))
			req := httptest.NewRequest(http.MethodPost, "/api/badges", bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			h.CreateBadge(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.assertBody != nil {
				tc.assertBody(t, rec)
			}
		})
	}
}

func TestGetBadgeHandler(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC()
	tests := []struct {
		name       string
		pathID     string
		reqHeaders map[string]string
		repo       *fakeRepo
		wantStatus int
		assertBody func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{name: "rejects invalid id", pathID: "not-a-uuid", repo: &fakeRepo{}, wantStatus: http.StatusBadRequest},
		{
			name:   "returns not found",
			pathID: id.String(),
			repo: &fakeRepo{
				getFn: func(context.Context, uuid.UUID) (repository.Badge, error) { return repository.Badge{}, sql.ErrNoRows },
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "returns svg badge",
			pathID: id.String(),
			repo: &fakeRepo{getFn: func(context.Context, uuid.UUID) (repository.Badge, error) {
				return repository.Badge{
					ID:        id,
					Subject:   "build",
					Status:    "passing",
					Color:     "green",
					Style:     "flat",
					UpdatedAt: now,
				}, nil
			}},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "build")
				assert.NotEmpty(t, rec.Header().Get("ETag"))
			},
		},
		{
			name:       "returns not modified for etag",
			pathID:     id.String(),
			reqHeaders: map[string]string{"If-None-Match": `W/"` + id.String() + `-` + "123" + `"`},
			repo: &fakeRepo{getFn: func(context.Context, uuid.UUID) (repository.Badge, error) {
				return repository.Badge{
					ID:        id,
					Subject:   "build",
					Status:    "passing",
					Color:     "green",
					Style:     "flat",
					UpdatedAt: time.Unix(0, 123).UTC(),
				}, nil
			}},
			wantStatus: http.StatusNotModified,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, tc.repo, newTokens(t))
			req := httptest.NewRequest(http.MethodGet, "/api/badges/"+tc.pathID, nil)
			req.SetPathValue("id", tc.pathID)
			for key, value := range tc.reqHeaders {
				req.Header.Set(key, value)
			}
			rec := httptest.NewRecorder()
			h.GetBadge(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.assertBody != nil {
				tc.assertBody(t, rec)
			}
		})
	}
}

func TestGetBadgeMetaHandler(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		pathID     string
		repo       *fakeRepo
		wantStatus int
		assertBody func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:   "returns badge meta",
			pathID: id.String(),
			repo: &fakeRepo{getFn: func(context.Context, uuid.UUID) (repository.Badge, error) {
				return repository.Badge{ID: id, Subject: "build", Status: "passing", Color: "green", Style: "flat"}, nil
			}},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp models.Badge
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, id.String(), resp.ID)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, tc.repo, newTokens(t))
			req := httptest.NewRequest(http.MethodGet, "/api/badges/"+tc.pathID+"/meta", nil)
			req.SetPathValue("id", tc.pathID)
			rec := httptest.NewRecorder()
			h.GetBadgeMeta(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.assertBody != nil {
				tc.assertBody(t, rec)
			}
		})
	}
}

func TestPatchBadgeHandler(t *testing.T) {
	id := uuid.New()
	token := "token"
	tokens := newTokens(t)
	hash, err := tokens.HashToken(token)
	require.NoError(t, err)

	tests := []struct {
		name       string
		body       string
		authHeader string
		handler    *handler.Handler
		wantStatus int
		assertBody func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "updates badge",
			body:       `{"subject":"updated"}`,
			authHeader: "Bearer " + token,
			handler: newHandler(t, &fakeRepo{
				getFn: func(context.Context, uuid.UUID) (repository.Badge, error) {
					return repository.Badge{
						ID:        id,
						TokenHash: hash,
						Subject:   "build",
						Status:    "passing",
						Color:     "green",
						Style:     "flat",
					}, nil
				},
				updateFn: func(_ context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error) {
					return repository.Badge{
						ID:        id,
						TokenHash: hash,
						Subject:   arg.Subject,
						Status:    arg.Status,
						Color:     arg.Color,
						Style:     arg.Style,
					}, nil
				},
			}, tokens),
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp models.Badge
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.Equal(t, "updated", resp.Subject)
			},
		},
		{
			name:       "requires token",
			body:       `{"subject":"updated"}`,
			handler:    &handler.Handler{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "requires fields",
			body:       `{}`,
			authHeader: "Bearer token",
			handler:    &handler.Handler{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/badges/"+id.String(), bytes.NewBufferString(tc.body))
			req.SetPathValue("id", id.String())
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			tc.handler.PatchBadge(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.assertBody != nil {
				tc.assertBody(t, rec)
			}
		})
	}
}

func TestDeleteBadgeHandler(t *testing.T) {
	id := uuid.New()
	token := "token"
	tokens := newTokens(t)
	hash, err := tokens.HashToken(token)
	require.NoError(t, err)

	tests := []struct {
		name       string
		authHeader string
		handler    *handler.Handler
		wantStatus int
	}{
		{
			name:       "deletes badge",
			authHeader: "Bearer " + token,
			handler: newHandler(t, &fakeRepo{
				getFn: func(context.Context, uuid.UUID) (repository.Badge, error) {
					return repository.Badge{
						ID:        id,
						TokenHash: hash,
						Subject:   "build",
						Status:    "passing",
						Color:     "green",
						Style:     "flat",
					}, nil
				},
				deleteFn: func(context.Context, uuid.UUID) error { return nil },
			}, tokens),
			wantStatus: http.StatusNoContent,
		},
		{name: "requires token", handler: &handler.Handler{}, wantStatus: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/api/badges/"+id.String(), nil)
			req.SetPathValue("id", id.String())
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			tc.handler.DeleteBadge(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
		})
	}
}

func TestLiveBadgeHandler(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
		assertBody func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:       "rejects invalid input",
			url:        "/api/badges/live?status=passing&color=green",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "renders live badge",
			url:        "/api/badges/live?subject=build&status=passing&color=green",
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.Contains(t, rec.Body.String(), "build")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(t, &fakeRepo{}, newTokens(t))
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.LiveBadge(rec, req)
			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.assertBody != nil {
				tc.assertBody(t, rec)
			}
		})
	}
}
