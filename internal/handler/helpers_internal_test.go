package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhajizada/signum/internal/service"
)

func TestWriteServiceError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := &Handler{logger: logger}

	cases := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{
			name:   "invalid-input",
			err:    service.ErrInvalidBadgeInput,
			status: http.StatusBadRequest,
			body:   service.ErrInvalidBadgeInput.Error(),
		},
		{
			name:   "unauthorized",
			err:    service.ErrUnauthorized,
			status: http.StatusUnauthorized,
			body:   service.ErrUnauthorized.Error(),
		},
		{name: "not-found", err: service.ErrNotFound, status: http.StatusNotFound, body: service.ErrNotFound.Error()},
		{
			name:   "unknown",
			err:    io.ErrUnexpectedEOF,
			status: http.StatusInternalServerError,
			body:   "internal server error",
		},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		h.writeServiceError(rec, tc.err)
		if rec.Code != tc.status {
			t.Fatalf("%s: expected status %d, got %d", tc.name, tc.status, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), tc.body) {
			t.Fatalf("%s: expected body %q, got %q", tc.name, tc.body, rec.Body.String())
		}
	}
}
