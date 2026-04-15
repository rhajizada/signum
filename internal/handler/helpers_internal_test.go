package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhajizada/signum/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestWriteServiceError(t *testing.T) {
	h := &Handler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{
			name:   "invalid input",
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
		{name: "not found", err: service.ErrNotFound, status: http.StatusNotFound, body: service.ErrNotFound.Error()},
		{
			name:   "unknown",
			err:    io.ErrUnexpectedEOF,
			status: http.StatusInternalServerError,
			body:   "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.writeServiceError(rec, tc.err)
			assert.Equal(t, tc.status, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.body)
		})
	}
}
