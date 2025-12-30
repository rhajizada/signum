package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhajizada/signum/internal/config"
	"github.com/rhajizada/signum/internal/middleware"
)

func TestRateLimitAllowsNonAPI(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		Burst:             1,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := middleware.RateLimit(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/assets/logo.png", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for non-api route, got %d", rec.Code)
	}
}

func TestRateLimitBlocksAfterBurst(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		Burst:             1,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := middleware.RateLimit(cfg)(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/badges", nil)
	req.RemoteAddr = "10.0.0.1:1234"

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected first request ok, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit, got %d", rec.Code)
	}
}

func TestRateLimitIsPerIP(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		Burst:             1,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := middleware.RateLimit(cfg)(handler)

	reqA := httptest.NewRequest(http.MethodPost, "/api/badges", nil)
	reqA.RemoteAddr = "10.0.0.1:1234"
	reqB := httptest.NewRequest(http.MethodPost, "/api/badges", nil)
	reqB.RemoteAddr = "10.0.0.2:1234"

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, reqA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for ip A, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, reqB)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for ip B, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, reqA)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected limit for ip A, got %d", rec.Code)
	}
}

func TestRateLimitSkipsLiveBadge(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		Burst:             1,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := middleware.RateLimit(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/live", nil)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for live badge, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for live badge, got %d", rec.Code)
	}
}

func TestRateLimitSkipsStoredBadge(t *testing.T) {
	cfg := config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		Burst:             1,
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := middleware.RateLimit(cfg)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/badges/123", nil)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for stored badge, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok for stored badge, got %d", rec.Code)
	}
}
