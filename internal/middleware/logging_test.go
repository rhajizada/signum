package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

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
	for k, v := range h.lastAttrs {
		out[k] = v
	}
	return out
}

func TestWrappedWriterWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	writer := &wrappedWriter{ResponseWriter: rec}

	writer.WriteHeader(http.StatusAccepted)
	if writer.statusCode != http.StatusAccepted {
		t.Fatalf("expected status to be %d, got %d", http.StatusAccepted, writer.statusCode)
	}
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected recorder status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestExtractParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/badges/abc", nil)
	req.SetPathValue("id", "abc")

	params := extractParams("GET /api/badges/{id}", req)
	if params == nil || params["id"] != "abc" {
		t.Fatalf("expected params to include id=abc, got %#v", params)
	}

	if got := extractParams("GET /api/badges/static", req); got != nil {
		t.Fatalf("expected nil params for static pattern, got %#v", got)
	}

	if got := extractParams("", req); got != nil {
		t.Fatalf("expected nil params for empty pattern, got %#v", got)
	}

	if got := extractParams("GET /api/badges/{id}", nil); got != nil {
		t.Fatalf("expected nil params for nil request, got %#v", got)
	}
}

func TestLoggingAddsAttrs(t *testing.T) {
	capture := &captureHandler{}
	logger := slog.New(capture)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestctx.WithRoutePattern(r.Context(), "GET /things/{id}")
		r.SetPathValue("id", "123")
		w.WriteHeader(http.StatusAccepted)
	})

	mw := Logging(logger)(handler)
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
