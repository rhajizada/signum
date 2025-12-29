package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rhajizada/signum/pkg/renderer"
)

var ErrInvalidBadgeInput = errors.New("invalid badge input")

func (s *Service) GetLiveBadge(subject, status, color, style string) ([]byte, error) {
	if s == nil || s.r == nil {
		return nil, errors.New("renderer is not configured")
	}

	subject = strings.TrimSpace(subject)
	status = strings.TrimSpace(status)
	color = strings.TrimSpace(color)
	style = strings.TrimSpace(style)

	if subject == "" {
		return nil, fmt.Errorf("%w: subject is required", ErrInvalidBadgeInput)
	}
	if status == "" {
		return nil, fmt.Errorf("%w: status is required", ErrInvalidBadgeInput)
	}
	if color == "" {
		return nil, fmt.Errorf("%w: color is required", ErrInvalidBadgeInput)
	}
	if style == "" {
		style = string(renderer.StyleFlat)
	}

	badgeColor := renderer.Color(color)
	if !badgeColor.IsValid() {
		return nil, fmt.Errorf("%w: invalid color %q", ErrInvalidBadgeInput, color)
	}

	badgeStyle := renderer.Style(style)
	if !badgeStyle.IsValid() {
		return nil, fmt.Errorf("%w: invalid style %q", ErrInvalidBadgeInput, style)
	}

	return s.r.Render(renderer.Badge{
		Subject: subject,
		Status:  status,
		Color:   badgeColor,
		Style:   badgeStyle,
	})
}
