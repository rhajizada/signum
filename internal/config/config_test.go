package config_test

import (
	"testing"

	"github.com/rhajizada/signum/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresConfigDSN(t *testing.T) {
	tests := []struct {
		name     string
		config   config.PostgresConfig
		expected string
	}{
		{
			name: "builds dsn",
			config: config.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "user",
				Password: "pass",
				DBName:   "signum",
				SSLMode:  "require",
			},
			expected: "postgres://user:pass@localhost:5432/signum?sslmode=require",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.config.DSN())
		})
	}
}

func TestLoadServer(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		assert func(t *testing.T, cfg *config.ServerConfig, err error)
	}{
		{
			name: "loads from env",
			env: map[string]string{
				"SIGNUM_ADDR":                           ":9090",
				"SIGNUM_FONT_PATH":                      "/tmp/font.ttf",
				"SIGNUM_SECRET_KEY":                     "secret",
				"SIGNUM_POSTGRES_HOST":                  "db",
				"SIGNUM_POSTGRES_PORT":                  "1234",
				"SIGNUM_POSTGRES_USER":                  "pguser",
				"SIGNUM_POSTGRES_PASSWORD":              "pgpass",
				"SIGNUM_POSTGRES_DBNAME":                "signum",
				"SIGNUM_POSTGRES_SSLMODE":               "verify-full",
				"SIGNUM_RATE_LIMIT_ENABLED":             "true",
				"SIGNUM_RATE_LIMIT_REQUESTS_PER_MINUTE": "120",
				"SIGNUM_RATE_LIMIT_BURST":               "40",
			},
			assert: func(t *testing.T, cfg *config.ServerConfig, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				assert.Equal(t, ":9090", cfg.Address)
				assert.Equal(t, "/tmp/font.ttf", cfg.FontPath)
				assert.Equal(t, "secret", cfg.SecretKey)
				assert.Equal(t, 1234, cfg.Postgres.Port)
				assert.Equal(t, "verify-full", cfg.Postgres.SSLMode)
				assert.True(t, cfg.RateLimit.Enabled)
				assert.Equal(t, 120, cfg.RateLimit.RequestsPerMinute)
				assert.Equal(t, 40, cfg.RateLimit.Burst)
			},
		},
		{
			name: "rejects invalid env",
			env: map[string]string{
				"SIGNUM_ADDR":              ":9090",
				"SIGNUM_FONT_PATH":         "/tmp/font.ttf",
				"SIGNUM_SECRET_KEY":        "secret",
				"SIGNUM_POSTGRES_HOST":     "db",
				"SIGNUM_POSTGRES_PORT":     "not-a-number",
				"SIGNUM_POSTGRES_USER":     "pguser",
				"SIGNUM_POSTGRES_PASSWORD": "pgpass",
				"SIGNUM_POSTGRES_DBNAME":   "signum",
			},
			assert: func(t *testing.T, cfg *config.ServerConfig, err error) {
				require.Error(t, err)
				assert.Nil(t, cfg)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			cfg, err := config.LoadServer()
			tc.assert(t, cfg, err)
		})
	}
}
