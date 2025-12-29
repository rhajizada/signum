package config

import (
	"net/url"
	"strconv"

	"github.com/caarlos0/env/v11"
)

// ServerConfig holds every runtime option for the HTTP server.
type ServerConfig struct {
	Address   string         `env:"SIGNUM_ADDR"       envDefault:":8080"`
	Postgres  PostgresConfig `                                           envPrefix:"SIGNUM_POSTGRES_"`
	FontPath  string         `env:"SIGNUM_FONT_PATH"                                                  envRequired:"true"`
	SecretKey string         `env:"SIGNUM_SECRET_KEY"                                                 envRequired:"true"`
}

// PostgresConfig holds database connection settings.
type PostgresConfig struct {
	Host     string `env:"HOST"     envRequired:"true"`
	Port     int    `env:"PORT"                        envDefault:"5432"`
	User     string `env:"USER"     envRequired:"true"`
	Password string `env:"PASSWORD" envRequired:"true"`
	DBName   string `env:"DBNAME"   envRequired:"true"`
	SSLMode  string `env:"SSLMODE"                     envDefault:"disable"`
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
