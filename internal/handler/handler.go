package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

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

// LiveBadge godoc
//
//	@Summary		Render a live badge
//	@Description	Renders an SVG badge for the provided parameters.
//	@Tags			Badges
//	@Produce		image/svg+xml
//	@Param			subject	query		string	true	"Left-hand subject text"
//	@Param			status	query		string	true	"Right-hand status text"
//	@Param			color	query		string	true	"Badge color (named or hex)"
//	@Param			style	query		string	false	"Badge style (flat, flat-square, plastic). Default: flat"
//	@Success		200		{string}	string	"SVG image"
//	@Failure		400		{string}	string
//	@Failure		500		{string}	string
//	@Router			/api/badges/live [get].
func (h *Handler) LiveBadge(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	subject := query.Get("subject")
	status := query.Get("status")
	color := query.Get("color")
	style := query.Get("style")

	badge, err := h.svc.GetLiveBadge(subject, status, color, style)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidBadgeInput) {
			statusCode = http.StatusBadRequest
		}
		writeError(w, statusCode, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(badge)
}
