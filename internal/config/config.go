package config

import (
	"net/url"
	"strconv"

	"github.com/caarlos0/env/v11"
)

// ServerConfig holds every runtime option for the HTTP server.
type ServerConfig struct {
	Address   string `env:"SIGNUM_ADDR"       envDefault:":8080"`
	Postgres  PostgresConfig
	FontPath  string `env:"SIGNUM_FONT_PATH"                     envRequired:"true"`
	SecretKey string `env:"SIGNUM_SECRET_KEY"                    envRequired:"true"`
	RateLimit RateLimitConfig
}

// RateLimitConfig holds settings for API rate limiting.
type RateLimitConfig struct {
	Enabled           bool `env:"SIGNUM_RATE_LIMIT_ENABLED"             envDefault:"true"`
	RequestsPerMinute int  `env:"SIGNUM_RATE_LIMIT_REQUESTS_PER_MINUTE" envDefault:"20"`
	Burst             int  `env:"SIGNUM_RATE_LIMIT_BURST"               envDefault:"5"`
}

// PostgresConfig holds database connection settings.
type PostgresConfig struct {
	Host     string `env:"SIGNUM_POSTGRES_HOST"     envRequired:"true"`
	Port     int    `env:"SIGNUM_POSTGRES_PORT"                        envDefault:"5432"`
	User     string `env:"SIGNUM_POSTGRES_USER"     envRequired:"true"`
	Password string `env:"SIGNUM_POSTGRES_PASSWORD" envRequired:"true"`
	DBName   string `env:"SIGNUM_POSTGRES_DBNAME"   envRequired:"true"`
	SSLMode  string `env:"SIGNUM_POSTGRES_SSLMODE"                     envDefault:"disable"`
}

// DSN builds a Postgres connection string.
func (c PostgresConfig) DSN() string {
	escapedUser := url.UserPassword(c.User, c.Password)
	dsn := url.URL{
		Scheme: "postgres",
		User:   escapedUser,
		Host:   c.Host + ":" + strconv.Itoa(c.Port),
		Path:   c.DBName,
	}
	query := url.Values{}
	if c.SSLMode != "" {
		query.Set("sslmode", c.SSLMode)
	}
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

// LoadServer populates ServerConfig from environment variables.
func LoadServer() (*ServerConfig, error) {
	var cfg ServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
