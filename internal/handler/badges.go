package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rhajizada/signum/internal/models"
	"github.com/rhajizada/signum/internal/service"
)

const maxJSONBodyBytes int64 = 64 * 1024

// LiveBadge godoc
//
//	@Summary		Render a live badge
//	@Description	Renders an SVG badge for the provided parameters.
//	@Tags			Badges
//	@Produce		text/plain
//	@Param			subject	query		string	true	"Left-hand subject text"
//	@Param			status	query		string	true	"Right-hand status text"
//	@Param			color	query		string	true	"Badge color (named or hex)"
//	@Param			style	query		string	false	"Badge style (flat, flat-square, plastic). Default: flat"
//	@Success		200		{string}	string	"SVG image"
//	@Failure		400		{string}	string
//	@Failure		413		{string}	string
//	@Failure		429		{string}	string
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

// CreateBadge handles POST /api/badges.
//
//	@Summary		Create a badge
//	@Description	Stores a badge definition and returns its id and token.
//	@Tags			Badges
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		models.CreateBadgeRequest	true	"Create Badge Request"
//	@Success		201		{object}	models.CreateBadgeResponse
//	@Failure		400		{string}	string
//	@Failure		413		{string}	string
//	@Failure		429		{string}	string
//	@Failure		500		{string}	string
//	@Router			/api/badges [post].
func (h *Handler) CreateBadge(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, maxJSONBodyBytes)
	var payload models.CreateBadgeRequest
	if err := decodeJSON(req, &payload); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	badge, token, err := h.svc.CreateBadge(req.Context(), service.BadgeInput{
		Subject: payload.Subject,
		Status:  payload.Status,
		Color:   payload.Color,
		Style:   payload.Style,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, models.CreateBadgeResponse{
		Badge: toBadgeResponse(badge),
		Token: token,
	})
}

// GetBadge handles GET /api/badges/{id}.
//
//	@Summary		Render a stored badge
//	@Description	Returns an SVG badge for the stored definition.
//	@Tags			Badges
//	@Produce		text/plain
//	@Param			id	path		string	true	"Badge ID"
//	@Success		200	{string}	string	"SVG image"
//	@Failure		400	{string}	string
//	@Failure		404	{string}	string
//	@Failure		429	{string}	string
//	@Failure		500	{string}	string
//	@Router			/api/badges/{id} [get].
func (h *Handler) GetBadge(w http.ResponseWriter, req *http.Request) {
	id, err := parseBadgeID(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	badge, svg, err := h.svc.RenderBadge(req.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	etag := fmt.Sprintf(`W/"%s-%d"`, badge.ID, badge.UpdatedAt.UnixNano())
	lastModified := badge.UpdatedAt.UTC().Format(http.TimeFormat)
	if match := req.Header.Get("If-None-Match"); match != "" {
		if etagMatches(match, etag) {
			writeBadgeCacheHeaders(w, etag, lastModified)
			w.WriteHeader(http.StatusNotModified)
			return
		}
	} else if modifiedSince := req.Header.Get("If-Modified-Since"); modifiedSince != "" {
		if parsedTime, parseErr := time.Parse(http.TimeFormat, modifiedSince); parseErr == nil && !badge.UpdatedAt.After(parsedTime) {
			writeBadgeCacheHeaders(w, etag, lastModified)
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	writeBadgeCacheHeaders(w, etag, lastModified)
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svg)
}

// GetBadgeMeta handles GET /api/badges/{id}/meta.
//
//	@Summary		Read badge metadata
//	@Description	Returns the stored badge fields without the token.
//	@Tags			Badges
//	@Produce		json
//	@Param			id	path		string	true	"Badge ID"
//	@Success		200	{object}	models.Badge
//	@Failure		400	{string}	string
//	@Failure		404	{string}	string
//	@Failure		429	{string}	string
//	@Failure		500	{string}	string
//	@Router			/api/badges/{id}/meta [get].
func (h *Handler) GetBadgeMeta(w http.ResponseWriter, req *http.Request) {
	id, err := parseBadgeID(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	badge, err := h.svc.GetBadge(req.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBadgeResponse(badge))
}

// PatchBadge handles PATCH /api/badges/{id}.
//
//	@Summary		Patch a badge
//	@Description	Updates one or more fields in the stored badge definition.
//	@Tags			Badges
//	@Accept			json
//	@Produce		json
//	@Param			id				path	string	true	"Badge ID"
//	@Param			Authorization	header	string	true	"Token"
//	@Security		BearerAuth
//	@Param			payload	body		models.PatchBadgeRequest	true	"Patch Badge request"
//	@Success		200		{object}	models.Badge
//	@Failure		400		{string}	string
//	@Failure		401		{string}	string
//	@Failure		404		{string}	string
//	@Failure		413		{string}	string
//	@Failure		429		{string}	string
//	@Failure		500		{string}	string
//	@Router			/api/badges/{id} [patch].
func (h *Handler) PatchBadge(w http.ResponseWriter, req *http.Request) {
	id, err := parseBadgeID(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token := readBearerToken(req)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, maxJSONBodyBytes)
	var payload models.PatchBadgeRequest
	err = decodeJSON(req, &payload)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if payload.Subject == nil && payload.Status == nil && payload.Color == nil && payload.Style == nil {
		writeError(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	badge, err := h.svc.PatchBadge(req.Context(), id, token, service.BadgePatch{
		Subject: payload.Subject,
		Status:  payload.Status,
		Color:   payload.Color,
		Style:   payload.Style,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toBadgeResponse(badge))
}

// DeleteBadge handles DELETE /api/badges/{id}.
//
//	@Summary		Delete a badge
//	@Description	Deletes the stored badge definition.
//	@Tags			Badges
//	@Param			id				path	string	true	"Badge ID"
//	@Param			Authorization	header	string	true	"Token"
//	@Security		BearerAuth
//	@Success		204	{string}	string
//	@Failure		400	{string}	string
//	@Failure		401	{string}	string
//	@Failure		404	{string}	string
//	@Failure		429	{string}	string
//	@Failure		500	{string}	string
//	@Router			/api/badges/{id} [delete].
func (h *Handler) DeleteBadge(w http.ResponseWriter, req *http.Request) {
	id, err := parseBadgeID(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token := readBearerToken(req)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	err = h.svc.DeleteBadge(req.Context(), id, token)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(req *http.Request, dst any) error {
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("invalid JSON payload")
	}
	return nil
}

func parseBadgeID(req *http.Request) (uuid.UUID, error) {
	id := strings.TrimSpace(req.PathValue("id"))
	if id == "" {
		return uuid.UUID{}, errors.New("badge id is required")
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, errors.New("invalid badge id")
	}
	return parsed, nil
}

func readBearerToken(req *http.Request) string {
	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(auth, prefix))
}

func writeBadgeCacheHeaders(w http.ResponseWriter, etag, lastModified string) {
	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, must-revalidate")
	w.Header().Set("ETag", etag)
	w.Header().Set("Last-Modified", lastModified)
}

func etagMatches(header, etag string) bool {
	for part := range strings.SplitSeq(header, ",") {
		if strings.TrimSpace(part) == etag {
			return true
		}
	}
	return false
}

func toBadgeResponse(badge service.Badge) models.Badge {
	return models.Badge{
		ID:        badge.ID.String(),
		Subject:   badge.Subject,
		Status:    badge.Status,
		Color:     badge.Color,
		Style:     badge.Style,
		CreatedAt: badge.CreatedAt,
		UpdatedAt: badge.UpdatedAt,
	}
}
