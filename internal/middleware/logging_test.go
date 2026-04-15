package middleware_test

import (
	"context"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/rhajizada/signum/internal/middleware"
	"github.com/rhajizada/signum/internal/requestctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captureHandler struct {
	mu        sync.Mutex
	lastAttrs map[string]any
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.mu.Lock()
	h.lastAttrs = attrs
	h.mu.Unlock()
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler        { return h }

func (h *captureHandler) snapshot() map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make(map[string]any, len(h.lastAttrs))
	maps.Copy(out, h.lastAttrs)
	return out
}

func TestLoggingAddsAttrs(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		route         string
		pathParam     string
		wantStatus    int
		wantRoute     string
		wantParamName string
		wantParamVal  string
	}{
		{
			name:          "captures route attrs",
			path:          "/things/123",
			route:         "GET /things/{id}",
			pathParam:     "123",
			wantStatus:    http.StatusAccepted,
			wantRoute:     "GET /things/{id}",
			wantParamName: "id",
			wantParamVal:  "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			capture := &captureHandler{}
			logger := slog.New(capture)
			mw := middleware.Logging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestctx.WithRoutePattern(r.Context(), tc.wantRoute)
				r.SetPathValue(tc.wantParamName, tc.pathParam)
				w.WriteHeader(tc.wantStatus)
			}))

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, req)

			attrs := capture.snapshot()
			require.NotNil(t, attrs)
			assert.Equal(t, int64(tc.wantStatus), attrs["status"])
			assert.Equal(t, tc.wantRoute, attrs["route"])
			assert.Equal(t, map[string]string{tc.wantParamName: tc.wantParamVal}, attrs["params"])
		})
	}
}
