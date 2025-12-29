package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/pkg/renderer"
)

// Service contains the business logic that interacts with persistence.
type Service struct {
	r      *renderer.Renderer
	repo   BadgeRepository
	tokens *TokenManager
}

// New creates a Service instance.
func New(r *renderer.Renderer, repo BadgeRepository, tokens *TokenManager) (*Service, error) {
	if r == nil {
		return nil, errors.New("renderer is required")
	}
	if repo == nil {
		return nil, errors.New("repository is required")
	}
	if tokens == nil {
		return nil, errors.New("token manager is required")
	}
	return &Service{
		r:      r,
		repo:   repo,
		tokens: tokens,
	}, nil
}

// BadgeRepository defines the data access needed by the service layer.
type BadgeRepository interface {
	CreateBadge(ctx context.Context, arg repository.CreateBadgeParams) (repository.Badge, error)
	GetBadgeByID(ctx context.Context, id uuid.UUID) (repository.Badge, error)
	UpdateBadge(ctx context.Context, arg repository.UpdateBadgeParams) (repository.Badge, error)
	DeleteBadge(ctx context.Context, id uuid.UUID) error
}
