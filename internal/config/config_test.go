package config_test

import (
	"testing"

	"github.com/rhajizada/signum/internal/config"
)

func TestPostgresConfigDSN(t *testing.T) {
	cfg := config.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "pass",
		DBName:   "signum",
		SSLMode:  "require",
	}

	got := cfg.DSN()
	expected := "postgres://user:pass@localhost:5432/signum?sslmode=require"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestLoadServerFromEnv(t *testing.T) {
	t.Setenv("SIGNUM_ADDR", ":9090")
	t.Setenv("SIGNUM_FONT_PATH", "/tmp/font.ttf")
	t.Setenv("SIGNUM_SECRET_KEY", "secret")
	t.Setenv("SIGNUM_POSTGRES_HOST", "db")
	t.Setenv("SIGNUM_POSTGRES_PORT", "1234")
	t.Setenv("SIGNUM_POSTGRES_USER", "pguser")
	t.Setenv("SIGNUM_POSTGRES_PASSWORD", "pgpass")
	t.Setenv("SIGNUM_POSTGRES_DBNAME", "signum")
	t.Setenv("SIGNUM_POSTGRES_SSLMODE", "verify-full")

	cfg, err := config.LoadServer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Address != ":9090" {
		t.Fatalf("expected address :9090, got %q", cfg.Address)
	}
	if cfg.FontPath != "/tmp/font.ttf" {
		t.Fatalf("expected font path, got %q", cfg.FontPath)
	}
	if cfg.SecretKey != "secret" {
		t.Fatalf("expected secret key to be set")
	}
	if cfg.Postgres.Port != 1234 {
		t.Fatalf("expected postgres port 1234, got %d", cfg.Postgres.Port)
	}
	if cfg.Postgres.SSLMode != "verify-full" {
		t.Fatalf("expected sslmode verify-full, got %q", cfg.Postgres.SSLMode)
	}
}
