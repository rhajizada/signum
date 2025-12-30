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
)

type captureHandler struct {
	mu        sync.Mutex
	lastAttrs map[string]any
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

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

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *captureHandler) WithGroup(string) slog.Handler {
	return h
}

func (h *captureHandler) snapshot() map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make(map[string]any, len(h.lastAttrs))
	maps.Copy(out, h.lastAttrs)
	return out
}

func TestLoggingAddsAttrs(t *testing.T) {
	capture := &captureHandler{}
	logger := slog.New(capture)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestctx.WithRoutePattern(r.Context(), "GET /things/{id}")
		r.SetPathValue("id", "123")
		w.WriteHeader(http.StatusAccepted)
	})

	mw := middleware.Logging(logger)(handler)
	req := httptest.NewRequest(http.MethodGet, "/things/123", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	attrs := capture.snapshot()
	statusAttr := attrs["status"]
	status := -1
	switch v := statusAttr.(type) {
	case int:
		status = v
	case int64:
		status = int(v)
	case float64:
		status = int(v)
	}
	if status != http.StatusAccepted {
		t.Fatalf("expected status attr %d, got %#v", http.StatusAccepted, attrs["status"])
	}
	if attrs["route"] != "GET /things/{id}" {
		t.Fatalf("expected route attr to be set, got %#v", attrs["route"])
	}
	params, ok := attrs["params"].(map[string]string)
	if !ok || params["id"] != "123" {
		t.Fatalf("expected params attr to include id=123, got %#v", attrs["params"])
	}
}
