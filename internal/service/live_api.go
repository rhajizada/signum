package service

import (
	"errors"

	"github.com/rhajizada/signum/pkg/renderer"
)

var ErrInvalidBadgeInput = errors.New("invalid badge input")

func (s *Service) GetLiveBadge(subject, status, color, style string) ([]byte, error) {
	if s == nil || s.r == nil {
		return nil, errors.New("renderer is not configured")
	}

	var err error
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
