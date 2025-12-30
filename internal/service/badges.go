package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rhajizada/signum/internal/repository"
	"github.com/rhajizada/signum/pkg/renderer"
)

// Badge represents a stored badge definition.
type Badge struct {
	ID        uuid.UUID `json:"id"`
	Subject   string    `json:"subject"`
	Status    string    `json:"status"`
	Color     string    `json:"color"`
	Style     string    `json:"style"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BadgeInput is used for create and full updates.
type BadgeInput struct {
	Subject string
	Status  string
	Color   string
	Style   string
}

// BadgePatch is used for partial updates.
type BadgePatch struct {
	Subject *string
	Status  *string
	Color   *string
	Style   *string
}

var (
	ErrNotFound     = errors.New("badge not found")
	ErrUnauthorized = errors.New("unauthorized")
)

// CreateBadge stores a new badge definition and returns the token.
func (s *Service) CreateBadge(ctx context.Context, input BadgeInput) (Badge, string, error) {
	if s == nil || s.repo == nil || s.tokens == nil {
		return Badge{}, "", errors.New("service is not configured")
	}

	subject, status, color, style, err := normalizeBadgeInput(input.Subject, input.Status, input.Color, input.Style)
	if err != nil {
		return Badge{}, "", err
	}

	token, hash, err := s.tokens.GenerateToken()
	if err != nil {
		return Badge{}, "", err
	}

	row, err := s.repo.CreateBadge(ctx, repository.CreateBadgeParams{
		TokenHash: hash,
		Subject:   subject,
		Status:    status,
		Color:     color,
		Style:     style,
	})
	if err != nil {
		return Badge{}, "", err
	}

	return toBadge(row), token, nil
}

// GetBadge fetches a badge by id.
func (s *Service) GetBadge(ctx context.Context, id uuid.UUID) (Badge, error) {
	row, err := s.repo.GetBadgeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Badge{}, ErrNotFound
		}
		return Badge{}, err
	}
	return toBadge(row), nil
}

// RenderBadge renders a stored badge.
func (s *Service) RenderBadge(ctx context.Context, id uuid.UUID) ([]byte, error) {
	badge, err := s.GetBadge(ctx, id)
	if err != nil {
		return nil, err
	}

	subject := badge.Subject
	status := badge.Status
	color := badge.Color
	style := badge.Style

	subject, status, color, style, err = normalizeBadgeInput(subject, status, color, style)
	if err != nil {
		return nil, err
	}

	return s.r.Render(renderer.Badge{
		Subject: subject,
		Status:  status,
		Color:   renderer.Color(color),
		Style:   renderer.Style(style),
	})
}

// PatchBadge partially updates a badge definition after validating the token.
func (s *Service) PatchBadge(ctx context.Context, id uuid.UUID, token string, patch BadgePatch) (Badge, error) {
	if token == "" {
		return Badge{}, ErrUnauthorized
	}

	current, err := s.authorize(ctx, id, token)
	if err != nil {
		return Badge{}, err
	}

	subject := current.Subject
	status := current.Status
	color := current.Color
	style := current.Style

	if patch.Subject != nil {
		subject = strings.TrimSpace(*patch.Subject)
	}
	if patch.Status != nil {
		status = strings.TrimSpace(*patch.Status)
	}
	if patch.Color != nil {
		color = strings.TrimSpace(*patch.Color)
	}
	if patch.Style != nil {
		style = strings.TrimSpace(*patch.Style)
	}

	subject, status, color, style, err = normalizeBadgeInput(subject, status, color, style)
	if err != nil {
		return Badge{}, err
	}

	row, err := s.repo.UpdateBadge(ctx, repository.UpdateBadgeParams{
		ID:      id,
		Subject: subject,
		Status:  status,
		Color:   color,
		Style:   style,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Badge{}, ErrNotFound
		}
		return Badge{}, err
	}

	return toBadge(row), nil
}

// DeleteBadge removes a badge definition after validating the token.
func (s *Service) DeleteBadge(ctx context.Context, id uuid.UUID, token string) error {
	if token == "" {
		return ErrUnauthorized
	}

	if _, err := s.authorize(ctx, id, token); err != nil {
		return err
	}

	return s.repo.DeleteBadge(ctx, id)
}

func (s *Service) authorize(ctx context.Context, id uuid.UUID, token string) (repository.Badge, error) {
	row, err := s.repo.GetBadgeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repository.Badge{}, ErrNotFound
		}
		return repository.Badge{}, err
	}
	if !s.tokens.CompareHash(row.TokenHash, token) {
		return repository.Badge{}, ErrUnauthorized
	}
	return row, nil
}

func toBadge(row repository.Badge) Badge {
	return Badge{
		ID:        row.ID,
		Subject:   row.Subject,
		Status:    row.Status,
		Color:     row.Color,
		Style:     row.Style,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func normalizeBadgeInput(subject, status, color, style string) (string, string, string, string, error) {
	subject = strings.TrimSpace(subject)
	status = strings.TrimSpace(status)
	color = strings.TrimSpace(color)
	style = strings.TrimSpace(style)

	if subject == "" {
		return "", "", "", "", fmt.Errorf("%w: subject is required", ErrInvalidBadgeInput)
	}
	if status == "" {
		return "", "", "", "", fmt.Errorf("%w: status is required", ErrInvalidBadgeInput)
	}
	if color == "" {
		return "", "", "", "", fmt.Errorf("%w: color is required", ErrInvalidBadgeInput)
	}
	if style == "" {
		style = string(renderer.StyleFlat)
	}

	badgeColor := renderer.Color(color)
	if !badgeColor.IsValid() {
		return "", "", "", "", fmt.Errorf("%w: invalid color %q", ErrInvalidBadgeInput, color)
	}

	badgeStyle := renderer.Style(style)
	if !badgeStyle.IsValid() {
		return "", "", "", "", fmt.Errorf("%w: invalid style %q", ErrInvalidBadgeInput, style)
	}

	return subject, status, color, style, nil
}
