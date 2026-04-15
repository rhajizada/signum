package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueriesCRUD(t *testing.T) {
	type testCase struct {
		name string
		run  func(t *testing.T, queries *repository.Queries)
	}

	db := testutil.StartPostgres(t)
	queries := repository.New(db.DB)

	tests := []testCase{
		{
			name: "create and get badge",
			run: func(t *testing.T, queries *repository.Queries) {
				created, err := queries.CreateBadge(context.Background(), repository.CreateBadgeParams{
					TokenHash: "hash-1",
					Subject:   "build",
					Status:    "passing",
					Color:     "green",
					Style:     "flat",
				})
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, created.ID)
				assert.WithinDuration(t, time.Now(), created.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, created.CreatedAt, created.UpdatedAt, 5*time.Second)

				fetched, err := queries.GetBadgeByID(context.Background(), created.ID)
				require.NoError(t, err)
				assert.Equal(t, created.ID, fetched.ID)
				assert.Equal(t, "hash-1", fetched.TokenHash)
				assert.Equal(t, "build", fetched.Subject)
				assert.Equal(t, "passing", fetched.Status)
				assert.Equal(t, "green", fetched.Color)
				assert.Equal(t, "flat", fetched.Style)
			},
		},
		{
			name: "update badge",
			run: func(t *testing.T, queries *repository.Queries) {
				created, err := queries.CreateBadge(context.Background(), repository.CreateBadgeParams{
					TokenHash: "hash-2",
					Subject:   "build",
					Status:    "passing",
					Color:     "green",
					Style:     "flat",
				})
				require.NoError(t, err)

				updated, err := queries.UpdateBadge(context.Background(), repository.UpdateBadgeParams{
					ID:      created.ID,
					Subject: "deploy",
					Status:  "running",
					Color:   "blue",
					Style:   "plastic",
				})
				require.NoError(t, err)
				assert.Equal(t, created.ID, updated.ID)
				assert.Equal(t, created.TokenHash, updated.TokenHash)
				assert.Equal(t, "deploy", updated.Subject)
				assert.Equal(t, "running", updated.Status)
				assert.Equal(t, "blue", updated.Color)
				assert.Equal(t, "plastic", updated.Style)
				assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
			},
		},
		{
			name: "delete badge",
			run: func(t *testing.T, queries *repository.Queries) {
				created, err := queries.CreateBadge(context.Background(), repository.CreateBadgeParams{
					TokenHash: "hash-3",
					Subject:   "build",
					Status:    "passing",
					Color:     "green",
					Style:     "flat",
				})
				require.NoError(t, err)

				require.NoError(t, queries.DeleteBadge(context.Background(), created.ID))

				_, err = queries.GetBadgeByID(context.Background(), created.ID)
				require.ErrorIs(t, err, sql.ErrNoRows)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.DB.ExecContext(context.Background(), "TRUNCATE TABLE badges")
			require.NoError(t, err)
			tc.run(t, queries)
		})
	}
}
