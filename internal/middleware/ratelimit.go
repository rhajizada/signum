package middleware

import (
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rhajizada/signum/internal/config"
)

const (
	apiPrefix   = "/api/"
	badgePrefix = "/api/badges/"
)

const (
	clientTTL        = 10 * time.Minute
	cleanupInterval  = 1 * time.Minute
	secondsPerMinute = 60.0
)

type tokenBucket struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
	rate   float64
	burst  float64
}

func newTokenBucket(ratePerSecond float64, burst int) *tokenBucket {
	b := float64(burst)
	if b <= 0 {
		b = 1
	}
	return &tokenBucket{
		tokens: b,
		last:   time.Now(),
		rate:   ratePerSecond,
		burst:  b,
	}
}

func (b *tokenBucket) Allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.burst, b.tokens+elapsed*b.rate)
		b.last = now
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

type limiterStore struct {
	mu            sync.Mutex
	buckets       map[string]*tokenBucket
	lastSeen      map[string]time.Time
	lastCleanup   time.Time
	ratePerSecond float64
	burst         int
}

func newLimiterStore(ratePerSecond float64, burst int) *limiterStore {
	return &limiterStore{
		buckets:       make(map[string]*tokenBucket),
		lastSeen:      make(map[string]time.Time),
		lastCleanup:   time.Now(),
		ratePerSecond: ratePerSecond,
		burst:         burst,
	}
}

func (s *limiterStore) bucketFor(key string, now time.Time) *tokenBucket {
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.Sub(s.lastCleanup) >= cleanupInterval {
		for k, seen := range s.lastSeen {
			if now.Sub(seen) > clientTTL {
				delete(s.lastSeen, k)
				delete(s.buckets, k)
			}
		}
		s.lastCleanup = now
	}

	bucket, ok := s.buckets[key]
	if !ok {
		bucket = newTokenBucket(s.ratePerSecond, s.burst)
		s.buckets[key] = bucket
	}
	s.lastSeen[key] = now
	return bucket
}

func shouldSkipRateLimit(req *http.Request) bool {
	if req == nil {
		return false
	}
	if req.Method != http.MethodGet {
		return false
	}
	if strings.HasPrefix(req.URL.Path, "/api/docs/") {
		return true
	}
	if req.URL.Path == "/api/badges/live" {
		return true
	}
	if !strings.HasPrefix(req.URL.Path, badgePrefix) {
		return false
	}
	rest := strings.TrimPrefix(req.URL.Path, badgePrefix)
	return rest != "" && !strings.Contains(rest, "/")
}

func clientKey(req *http.Request) string {
	if req == nil {
		return ""
	}
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(req.RemoteAddr)
}

// RateLimit applies a per-IP token bucket limiter to API routes.
func RateLimit(cfg config.RateLimitConfig) func(http.Handler) http.Handler {
	if !cfg.Enabled || cfg.RequestsPerMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	ratePerSecond := float64(cfg.RequestsPerMinute) / secondsPerMinute
	store := newLimiterStore(ratePerSecond, cfg.Burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, apiPrefix) {
				next.ServeHTTP(w, r)
				return
			}
			if shouldSkipRateLimit(r) {
				next.ServeHTTP(w, r)
				return
			}
			key := clientKey(r)
			if key == "" {
				key = "unknown"
			}
			bucket := store.bucketFor(key, time.Now())
			if !bucket.Allow(time.Now()) {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
