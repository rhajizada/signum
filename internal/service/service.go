package service

import (
	"github.com/rhajizada/signum/pkg/renderer"
)

// Service contains the business logic that interacts with persistence.
type Service struct {
	r *renderer.Renderer
}

// New creates a Service instance.
func New(r *renderer.Renderer) *Service {
	return &Service{
		r: r,
	}
}
