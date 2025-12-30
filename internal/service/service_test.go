package service_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

func newRenderer(tb testing.TB) *renderer.Renderer {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	if err != nil {
		tb.Fatalf("new renderer: %v", err)
	}
	return r
}

func TestNewRequiresDeps(t *testing.T) {
	_, err := service.New(nil, &fakeRepo{}, &service.TokenManager{})
	if err == nil {
		t.Fatalf("expected renderer error")
	}
	_, err = service.New(newRenderer(t), nil, &service.TokenManager{})
	if err == nil {
		t.Fatalf("expected repository error")
	}
	_, err = service.New(newRenderer(t), &fakeRepo{}, nil)
	if err == nil {
		t.Fatalf("expected token manager error")
	}
}

func TestGetLiveBadgeNormalizesInput(t *testing.T) {
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(newRenderer(t), &fakeRepo{}, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	output, err := svc.GetLiveBadge(" subject ", " status ", " green ", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(output), "subject") {
		t.Fatalf("expected normalized subject in output")
	}
}

func TestCreateBadge(t *testing.T) {
	now := time.Now()
	expectedID := uuid.New()
	repo := &fakeRepo{
		createFn: func(_ context.Context, arg repository.CreateBadgeParams) (repository.Badge, error) {
			if arg.Subject != "subject" || arg.Status != "status" || arg.Color != "green" || arg.Style != "flat" {
				t.Fatalf("unexpected create params: %#v", arg)
			}
			if arg.TokenHash == "" {
				t.Fatalf("expected token hash")
			}
			return repository.Badge{
				ID:        expectedID,
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
	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	badge, token, err := svc.CreateBadge(context.Background(), service.BadgeInput{
		Subject: " subject ",
		Status:  " status ",
		Color:   " green ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token")
	}
	if badge.ID != expectedID || badge.Subject != "subject" || badge.Style != "flat" {
		t.Fatalf("unexpected badge: %#v", badge)
	}
}

func TestCreateBadgeUnconfigured(t *testing.T) {
	var svc *service.Service
	if _, _, err := svc.CreateBadge(context.Background(), service.BadgeInput{}); err == nil {
		t.Fatalf("expected error for unconfigured service")
	}
}

func TestGetBadgeNotFound(t *testing.T) {
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{}, sql.ErrNoRows
		},
	}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = svc.GetBadge(context.Background(), uuid.New())
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRenderBadgeWithOverrides(t *testing.T) {
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
	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	override := "custom"
	svg, err := svc.RenderBadge(context.Background(), id, service.BadgeOverrides{Subject: &override})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(svg), "custom") {
		t.Fatalf("expected override in svg")
	}
}

func TestPatchBadgeUnauthorized(t *testing.T) {
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(newRenderer(t), &fakeRepo{}, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	_, err = svc.PatchBadge(context.Background(), uuid.New(), "", service.BadgePatch{})
	if !errors.Is(err, service.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestPatchBadgeSuccess(t *testing.T) {
	token := "token"
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	hash, err := tokens.HashToken(token)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}

	now := time.Now()
	id := uuid.New()
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{
				ID:        id,
				TokenHash: hash,
				Subject:   "build",
				Status:    "passing",
				Color:     "green",
				Style:     "flat",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
		updateFn: func(_ context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error) {
			if arg.Subject != "updated" {
				t.Fatalf("expected updated subject, got %q", arg.Subject)
			}
			return repository.Badge{
				ID:        id,
				TokenHash: hash,
				Subject:   arg.Subject,
				Status:    arg.Status,
				Color:     arg.Color,
				Style:     arg.Style,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	subject := " updated "
	badge, err := svc.PatchBadge(context.Background(), id, token, service.BadgePatch{Subject: &subject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if badge.Subject != "updated" {
		t.Fatalf("unexpected badge: %#v", badge)
	}
}

func TestDeleteBadgeUnauthorized(t *testing.T) {
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(newRenderer(t), &fakeRepo{}, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	err = svc.DeleteBadge(context.Background(), uuid.New(), "")
	if !errors.Is(err, service.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestDeleteBadgeNotFound(t *testing.T) {
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	repo := &fakeRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (repository.Badge, error) {
			return repository.Badge{}, sql.ErrNoRows
		},
	}
	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	err = svc.DeleteBadge(context.Background(), uuid.New(), "token")
	if !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestDeleteBadge(t *testing.T) {
	token := "token"
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	hash, err := tokens.HashToken(token)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}

	id := uuid.New()
	deleted := false
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
		deleteFn: func(_ context.Context, got uuid.UUID) error {
			if got != id {
				t.Fatalf("unexpected id in delete")
			}
			deleted = true
			return nil
		},
	}
	svc, err := service.New(newRenderer(t), repo, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	err = svc.DeleteBadge(context.Background(), id, token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected delete to be called")
	}
}

func TestGetLiveBadge(t *testing.T) {
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	svc, err := service.New(newRenderer(t), &fakeRepo{}, tokens)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	output, err := svc.GetLiveBadge("build", "passing", "green", "flat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(output), "build") {
		t.Fatalf("expected subject in output")
	}
	_, err = svc.GetLiveBadge("", "passing", "green", "flat")
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestTokenManager(t *testing.T) {
	if _, err := service.NewTokenManager(""); err == nil {
		t.Fatalf("expected error for empty secret")
	}
	mgr, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	token, hash, err := mgr.GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" || hash == "" {
		t.Fatalf("expected token and hash")
	}
	_, err = mgr.HashToken("")
	if err == nil {
		t.Fatalf("expected error for empty token")
	}
	if !mgr.CompareHash(hash, token) {
		t.Fatalf("expected hash match")
	}
	if mgr.CompareHash(hash, "other") {
		t.Fatalf("expected hash mismatch")
	}
}

func TestTokenManagerNil(t *testing.T) {
	var mgr *service.TokenManager
	if _, _, err := mgr.GenerateToken(); err == nil {
		t.Fatalf("expected error for nil manager")
	}
	if _, err := mgr.HashToken("token"); err == nil {
		t.Fatalf("expected error for nil manager")
	}
	if mgr.CompareHash("hash", "token") {
		t.Fatalf("expected comparison to be false")
	}
}
