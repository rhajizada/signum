package config

import (
	"github.com/caarlos0/env/v11"
)

// ServerConfig holds every runtime option for the HTTP server.
type ServerConfig struct {
	Address  string `env:"SIGNUM_ADDR" envDefault:":8080"`
	FontPath string `env:"SIGNUM_FONT_PATH"`
}

// LoadServer populates ServerConfig from environment variables.
func LoadServer() (*ServerConfig, error) {
	var cfg ServerConfig
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
