package handler

import (
	"errors"
	"html/template"
	"log/slog"

	"github.com/rhajizada/signum/internal/service"
)

// Handler coordinates HTTP endpoints.
type Handler struct {
	svc    *service.Service
	logger *slog.Logger
	home   *template.Template
}

// New builds a Handler with the provided dependencies.
func New(svc *service.Service, logger *slog.Logger) (*Handler, error) {
	if svc == nil {
		return nil, errors.New("service is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	home, err := parseHomeTemplate()
	if err != nil {
		return nil, err
	}
	return &Handler{
		svc:    svc,
		logger: logger,
		home:   home,
	}, nil
}
