package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/service"
	"github.com/rhajizada/signum/internal/testutil"
	"github.com/rhajizada/signum/pkg/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font/basicfont"
)

type fakeRepo struct{}

func (f *fakeRepo) CreateBadge(context.Context, repository.CreateBadgeParams) (repository.Badge, error) {
	return repository.Badge{}, nil
}

func (f *fakeRepo) GetBadgeByID(context.Context, uuid.UUID) (repository.Badge, error) {
	return repository.Badge{}, sql.ErrNoRows
}

func (f *fakeRepo) UpdateBadge(context.Context, repository.UpdateBadgeParams) (repository.Badge, error) {
	return repository.Badge{}, nil
}

func (f *fakeRepo) DeleteBadge(context.Context, uuid.UUID) error {
	return nil
}

func newRenderer(tb testing.TB) *renderer.Renderer {
	tb.Helper()
	r, err := renderer.NewRendererWithFontFace(basicfont.Face7x13)
	require.NoError(tb, err)
	return r
}

func newTokenManager(tb testing.TB) *service.TokenManager {
	tb.Helper()
	tokens, err := service.NewTokenManager("secret")
	require.NoError(tb, err)
	return tokens
}

func newRepositoryService(
	tb testing.TB,
	queries service.BadgeRepository,
	tokens *service.TokenManager,
) *service.Service {
	tb.Helper()
	svc, err := service.New(newRenderer(tb), queries, tokens)
	require.NoError(tb, err)
	return svc
}

func TestNewRequiresDeps(t *testing.T) {
	tests := []struct {
		name      string
		renderer  *renderer.Renderer
		repo      service.BadgeRepository
		tokens    *service.TokenManager
		errString string
	}{
		{
			name:      "missing renderer",
			repo:      &fakeRepo{},
			tokens:    &service.TokenManager{},
			errString: "renderer is required",
		},
		{
			name:      "missing repository",
			renderer:  newRenderer(t),
			tokens:    &service.TokenManager{},
			errString: "repository is required",
		},
		{
			name:      "missing token manager",
			renderer:  newRenderer(t),
			repo:      &fakeRepo{},
			errString: "token manager is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := service.New(tc.renderer, tc.repo, tc.tokens)
			require.Nil(t, svc)
			require.EqualError(t, err, tc.errString)
		})
	}
}

func TestGetLiveBadge(t *testing.T) {
	tests := []struct {
		name       string
		subject    string
		status     string
		color      string
		style      string
		wantErr    error
		wantBody   string
		wantAbsent string
	}{
		{
			name:       "normalizes input and applies default style",
			subject:    " subject ",
			status:     " status ",
			color:      " green ",
			wantBody:   "subject",
			wantAbsent: " subject ",
		},
		{
			name:    "rejects missing subject",
			status:  "passing",
			color:   "green",
			style:   "flat",
			wantErr: service.ErrInvalidBadgeInput,
		},
	}

	svc := newRepositoryService(t, &fakeRepo{}, newTokenManager(t))

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := svc.GetLiveBadge(tc.subject, tc.status, tc.color, tc.style)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Contains(t, string(output), tc.wantBody)
			if tc.wantAbsent != "" {
				assert.NotContains(t, string(output), tc.wantAbsent)
			}
		})
	}
}

func TestBadgePersistenceFlow(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T, svc *service.Service, db *testutil.PostgresDB, tokens *service.TokenManager)
	}

	db := testutil.StartPostgres(t)
	repo := repository.New(db.DB)
	tokens := newTokenManager(t)
	svc := newRepositoryService(t, repo, tokens)

	tests := []testCase{
		{
			name: "create and get badge",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				badge, token, err := svc.CreateBadge(context.Background(), service.BadgeInput{
					Subject: " subject ",
					Status:  " status ",
					Color:   " green ",
				})
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotEqual(t, uuid.Nil, badge.ID)
				assert.Equal(t, "subject", badge.Subject)
				assert.Equal(t, "status", badge.Status)
				assert.Equal(t, "green", badge.Color)
				assert.Equal(t, "flat", badge.Style)

				stored, err := svc.GetBadge(context.Background(), badge.ID)
				require.NoError(t, err)
				assert.Equal(t, badge.ID, stored.ID)
				assert.Equal(t, badge.Subject, stored.Subject)
			},
		},
		{
			name: "patch badge",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				created, token, err := svc.CreateBadge(context.Background(), service.BadgeInput{
					Subject: "build",
					Status:  "passing",
					Color:   "green",
					Style:   "flat",
				})
				require.NoError(t, err)

				subject := " updated "
				updated, err := svc.PatchBadge(
					context.Background(),
					created.ID,
					token,
					service.BadgePatch{Subject: &subject},
				)
				require.NoError(t, err)
				assert.Equal(t, created.ID, updated.ID)
				assert.Equal(t, "updated", updated.Subject)
				assert.Equal(t, created.Status, updated.Status)
			},
		},
		{
			name: "delete badge",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				created, token, err := svc.CreateBadge(context.Background(), service.BadgeInput{
					Subject: "build",
					Status:  "passing",
					Color:   "green",
					Style:   "flat",
				})
				require.NoError(t, err)

				require.NoError(t, svc.DeleteBadge(context.Background(), created.ID, token))

				_, err = svc.GetBadge(context.Background(), created.ID)
				require.ErrorIs(t, err, service.ErrNotFound)
			},
		},
		{
			name: "render badge",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				created, _, err := svc.CreateBadge(context.Background(), service.BadgeInput{
					Subject: "build",
					Status:  "passing",
					Color:   "green",
					Style:   "flat",
				})
				require.NoError(t, err)

				badge, svg, err := svc.RenderBadge(context.Background(), created.ID)
				require.NoError(t, err)
				assert.Equal(t, created.ID, badge.ID)
				assert.Contains(t, string(svg), "build")
			},
		},
		{
			name: "render badge rejects invalid persisted input",
			run: func(t *testing.T, svc *service.Service, db *testutil.PostgresDB, tokens *service.TokenManager) {
				tokenHash, err := tokens.HashToken("token")
				require.NoError(t, err)

				var id uuid.UUID
				err = db.DB.QueryRowContext(
					context.Background(),
					`INSERT INTO badges (token_hash, subject, status, color, style) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
					tokenHash,
					"build",
					"passing",
					"not-a-color",
					"flat",
				).Scan(&id)
				require.NoError(t, err)

				_, _, err = svc.RenderBadge(context.Background(), id)
				require.ErrorIs(t, err, service.ErrInvalidBadgeInput)
			},
		},
		{
			name: "maps not found",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				_, err := svc.GetBadge(context.Background(), uuid.New())
				require.ErrorIs(t, err, service.ErrNotFound)

				err = svc.DeleteBadge(context.Background(), uuid.New(), "token")
				require.ErrorIs(t, err, service.ErrNotFound)
			},
		},
		{
			name: "rejects unauthorized patch and delete",
			run: func(t *testing.T, svc *service.Service, _ *testutil.PostgresDB, _ *service.TokenManager) {
				created, _, err := svc.CreateBadge(context.Background(), service.BadgeInput{
					Subject: "build",
					Status:  "passing",
					Color:   "green",
					Style:   "flat",
				})
				require.NoError(t, err)

				subject := "updated"
				_, err = svc.PatchBadge(context.Background(), created.ID, "", service.BadgePatch{Subject: &subject})
				require.ErrorIs(t, err, service.ErrUnauthorized)

				err = svc.DeleteBadge(context.Background(), created.ID, "")
				require.ErrorIs(t, err, service.ErrUnauthorized)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.DB.ExecContext(context.Background(), "TRUNCATE TABLE badges")
			require.NoError(t, err)
			tc.run(t, svc, db, tokens)
		})
	}
}

func TestServiceNilReceivers(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "create badge on nil service",
			run: func(t *testing.T) {
				var svc *service.Service
				_, _, err := svc.CreateBadge(context.Background(), service.BadgeInput{})
				require.EqualError(t, err, "service is not configured")
			},
		},
		{
			name: "token manager nil methods",
			run: func(t *testing.T) {
				var mgr *service.TokenManager
				_, _, err := mgr.GenerateToken()
				require.EqualError(t, err, "token manager is not configured")

				_, err = mgr.HashToken("token")
				require.EqualError(t, err, "token manager is not configured")
				assert.False(t, mgr.CompareHash("hash", "token"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestTokenManager(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects empty secret",
			run: func(t *testing.T) {
				mgr, err := service.NewTokenManager("")
				require.Nil(t, mgr)
				require.EqualError(t, err, "secret key is required")
			},
		},
		{
			name: "generates and compares tokens",
			run: func(t *testing.T) {
				mgr := newTokenManager(t)
				token, hash, err := mgr.GenerateToken()
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotEmpty(t, hash)
				assert.True(t, mgr.CompareHash(hash, token))
				assert.False(t, mgr.CompareHash(hash, "other"))
			},
		},
		{
			name: "rejects empty token hash input",
			run: func(t *testing.T) {
				mgr := newTokenManager(t)
				_, err := mgr.HashToken("")
				require.EqualError(t, err, "token is required")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}

func TestCreateBadgeTimestamps(t *testing.T) {
	db := testutil.StartPostgres(t)
	svc := newRepositoryService(t, repository.New(db.DB), newTokenManager(t))

	badge, _, err := svc.CreateBadge(context.Background(), service.BadgeInput{
		Subject: "build",
		Status:  "passing",
		Color:   "green",
		Style:   "flat",
	})
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), badge.CreatedAt, 5*time.Second)
	assert.WithinDuration(t, badge.CreatedAt, badge.UpdatedAt, 5*time.Second)
}
