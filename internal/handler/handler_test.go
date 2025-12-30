package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/models"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/pkg/renderer"
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

func newHandler(tb testing.TB, repo service.BadgeRepository, tokens *service.TokenManager) *Handler {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	if err != nil {
		tb.Fatalf("renderer: %v", err)
	}
	svc, err := service.New(r, repo, tokens)
	if err != nil {
		tb.Fatalf("service: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := New(svc, logger)
	if err != nil {
		tb.Fatalf("handler: %v", err)
	}
	return h
}

func TestWriteServiceError(t *testing.T) {
	h := &Handler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	rec := httptest.NewRecorder()
	h.writeServiceError(rec, service.ErrInvalidBadgeInput)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.writeServiceError(rec, service.ErrUnauthorized)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.writeServiceError(rec, service.ErrNotFound)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.writeServiceError(rec, errors.New("boom"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error, got %d", rec.Code)
	}
}

func TestNewHandlerRequiresService(t *testing.T) {
	if _, err := New(nil, nil); err == nil {
		t.Fatalf("expected error for nil service")
	}
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"subject":"a"}{"extra":"b"}`))
	var payload models.CreateBadgeRequest
	if err := decodeJSON(req, &payload); err == nil {
		t.Fatalf("expected trailing JSON error")
	}
}

func TestParseBadgeID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/badges/123", nil)
	if _, err := parseBadgeID(req); err == nil {
		t.Fatalf("expected missing id error")
	}
	req.SetPathValue("id", "not-a-uuid")
	if _, err := parseBadgeID(req); err == nil {
		t.Fatalf("expected invalid id error")
	}
}

func TestReadBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token := readBearerToken(req); token != "" {
		t.Fatalf("expected empty token")
	}
	req.Header.Set("Authorization", "Basic abc")
	if token := readBearerToken(req); token != "" {
		t.Fatalf("expected empty token for non-bearer")
	}
	req.Header.Set("Authorization", "Bearer abc")
	if token := readBearerToken(req); token != "abc" {
		t.Fatalf("expected token abc, got %q", token)
	}
}

func TestCreateBadgeHandler(t *testing.T) {
	now := time.Now()
	repo := &fakeRepo{
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
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	body := `{"subject":"build","status":"passing","color":"green","style":"flat"}`
	req := httptest.NewRequest(http.MethodPost, "/api/badges", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateBadge(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status created, got %d", rec.Code)
	}
	var resp models.CreateBadgeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token in response")
	}
	if resp.Subject != "build" || resp.Status != "passing" {
		t.Fatalf("unexpected badge response: %#v", resp.Badge)
	}
}

func TestCreateBadgeHandlerInvalidJSON(t *testing.T) {
	repo := &fakeRepo{}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodPost, "/api/badges", strings.NewReader(`{"subject":`))
	rec := httptest.NewRecorder()
	h.CreateBadge(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestGetBadgeHandlerWithOverride(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{
				ID:      id,
				Subject: "build",
				Status:  "passing",
				Color:   "green",
				Style:   "flat",
			}, nil
		},
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/"+id.String()+"?subject=custom", nil)
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.GetBadge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "custom") {
		t.Fatalf("expected override in svg response")
	}
}

func TestGetBadgeHandlerRejectsEmptyOverride(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/badges/123?subject=", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()
	h := &Handler{}
	h.GetBadge(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestGetBadgeHandlerNotFound(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{}, sql.ErrNoRows
		},
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.GetBadge(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got %d", rec.Code)
	}
}

func TestGetBadgeMetaHandler(t *testing.T) {
	id := uuid.New()
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{
				ID:      id,
				Subject: "build",
				Status:  "passing",
				Color:   "green",
				Style:   "flat",
			}, nil
		},
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/"+id.String()+"/meta", nil)
	req.SetPathValue("id", id.String())
	rec := httptest.NewRecorder()
	h.GetBadgeMeta(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	var resp models.Badge
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != id.String() {
		t.Fatalf("unexpected badge id: %s", resp.ID)
	}
}

func TestPatchBadgeHandler(t *testing.T) {
	id := uuid.New()
	token := "token"
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	hash, err := tokens.HashToken(token)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
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
	}
	h := newHandler(t, repo, tokens)

	payload := `{"subject":"updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/badges/"+id.String(), strings.NewReader(payload))
	req.SetPathValue("id", id.String())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.PatchBadge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	var resp models.Badge
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Subject != "updated" {
		t.Fatalf("expected updated subject, got %q", resp.Subject)
	}
}

func TestPatchBadgeHandlerMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/api/badges/123", strings.NewReader(`{"subject":"updated"}`))
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()
	h := &Handler{}
	h.PatchBadge(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rec.Code)
	}
}

func TestPatchBadgeHandlerMissingFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/api/badges/123", strings.NewReader(`{}`))
	req.SetPathValue("id", uuid.New().String())
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	h := &Handler{}
	h.PatchBadge(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestDeleteBadgeHandler(t *testing.T) {
	id := uuid.New()
	token := "token"
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	hash, err := tokens.HashToken(token)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{
				ID:        id,
				TokenHash: hash,
				Subject:   "build",
				Status:    "passing",
				Color:     "green",
				Style:     "flat",
			}, nil
		},
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodDelete, "/api/badges/"+id.String(), nil)
	req.SetPathValue("id", id.String())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.DeleteBadge(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected no content, got %d", rec.Code)
	}
}

func TestDeleteBadgeHandlerMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/badges/123", nil)
	req.SetPathValue("id", uuid.New().String())
	rec := httptest.NewRecorder()
	h := &Handler{}
	h.DeleteBadge(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", rec.Code)
	}
}

func TestLiveBadgeHandlerInvalidInput(t *testing.T) {
	repo := &fakeRepo{}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/live?status=passing&color=green", nil)
	rec := httptest.NewRecorder()
	h.LiveBadge(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", rec.Code)
	}
}

func TestLiveBadgeHandler(t *testing.T) {
	repo := &fakeRepo{}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/live?subject=build&status=passing&color=green", nil)
	rec := httptest.NewRecorder()
	h.LiveBadge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("build")) {
		t.Fatalf("expected svg response body")
	}
}
