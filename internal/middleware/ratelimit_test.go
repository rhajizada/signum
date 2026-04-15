package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhajizada/signum/internal/config"
	"github.com/rhajizada/signum/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimit(t *testing.T) {
	type requestSpec struct {
		path       string
		remoteAddr string
	}

	tests := []struct {
		name     string
		requests []requestSpec
		statuses []int
	}{
		{
			name:     "allows non api routes",
			requests: []requestSpec{{path: "/assets/logo.png"}},
			statuses: []int{http.StatusOK},
		},
		{
			name: "blocks after burst",
			requests: []requestSpec{
				{path: "/api/badges", remoteAddr: "10.0.0.1:1234"},
				{path: "/api/badges", remoteAddr: "10.0.0.1:1234"},
			},
			statuses: []int{http.StatusOK, http.StatusTooManyRequests},
		},
		{
			name: "tracks each ip independently",
			requests: []requestSpec{
				{path: "/api/badges", remoteAddr: "10.0.0.1:1234"},
				{path: "/api/badges", remoteAddr: "10.0.0.2:1234"},
				{path: "/api/badges", remoteAddr: "10.0.0.1:1234"},
			},
			statuses: []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests},
		},
		{
			name:     "skips live badge route",
			requests: []requestSpec{{path: "/api/badges/live"}, {path: "/api/badges/live"}},
			statuses: []int{http.StatusOK, http.StatusOK},
		},
		{
			name:     "skips stored badge route",
			requests: []requestSpec{{path: "/api/badges/123"}, {path: "/api/badges/123"}},
			statuses: []int{http.StatusOK, http.StatusOK},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.RateLimitConfig{Enabled: true, RequestsPerMinute: 1, Burst: 1}
			mw := middleware.RateLimit(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			for i, reqSpec := range tc.requests {
				method := http.MethodPost
				if reqSpec.path == "/api/badges/live" || reqSpec.path == "/api/badges/123" ||
					reqSpec.path == "/assets/logo.png" {
					method = http.MethodGet
				}
				req := httptest.NewRequest(method, reqSpec.path, nil)
				if reqSpec.remoteAddr != "" {
					req.RemoteAddr = reqSpec.remoteAddr
				}
				rec := httptest.NewRecorder()
				mw.ServeHTTP(rec, req)
				assert.Equal(t, tc.statuses[i], rec.Code)
			}
		})
	}
}
