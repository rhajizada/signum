package testutil

import (
	"context"
	"database/sql"
	"testing"
	"time"

	// Required for the pgx database/sql driver used by integration tests.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresStartupTimeout   = 2 * time.Minute
	postgresReadyOccurrences = 2
)

type PostgresDB struct {
	DB  *sql.DB
	DSN string
}

func StartPostgres(tb testing.TB) *PostgresDB {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), postgresStartupTimeout)
	tb.Cleanup(cancel)

	container, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("signum"),
		postgres.WithUsername("signum"),
		postgres.WithPassword("signum"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(postgresReadyOccurrences).
				WithStartupTimeout(time.Minute),
		),
	)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		require.NoError(tb, container.Terminate(context.Background()))
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(tb, err)

	db, err := sql.Open("pgx", dsn)
	require.NoError(tb, err)

	tb.Cleanup(func() {
		require.NoError(tb, db.Close())
	})

	require.NoError(tb, db.PingContext(ctx))
	require.NoError(tb, goose.SetDialect("postgres"))
	require.NoError(tb, goose.Up(db, "../../data/sql/migrations"))

	return &PostgresDB{DB: db, DSN: dsn}
}
