package requestctx_test

import (
	"context"
	"testing"

	"github.com/rhajizada/signum/internal/requestctx"
)

func TestEnsureAddsData(t *testing.T) {
	ctx := context.Background()
	ctx = requestctx.Ensure(ctx)
	if route, ok := requestctx.RoutePattern(ctx); ok || route != "" {
		t.Fatalf("expected no route pattern on ensured context")
	}
}

func TestWithRoutePatternAndBackendID(t *testing.T) {
	ctx := context.Background()

	ctx = requestctx.WithRoutePattern(ctx, "GET /api/badges/{id}")
	pattern, ok := requestctx.RoutePattern(ctx)
	if !ok || pattern != "GET /api/badges/{id}" {
		t.Fatalf("expected route pattern to be set, got %q (ok=%v)", pattern, ok)
	}

	ctx = requestctx.WithBackendID(ctx, "backend-1")
	backendID, ok := requestctx.BackendID(ctx)
	if !ok || backendID != "backend-1" {
		t.Fatalf("expected backend id to be set, got %q (ok=%v)", backendID, ok)
	}
}

func TestWithBackendIDEmptyNoop(t *testing.T) {
	var ctx context.Context
	ctx = requestctx.WithBackendID(ctx, "")
	if ctx != nil {
		t.Fatalf("expected nil context to remain nil")
	}
	if _, ok := requestctx.BackendID(context.Background()); ok {
		t.Fatalf("expected no backend id for nil context")
	}
}

func TestWithRoutePatternNilContext(t *testing.T) {
	ctx := requestctx.WithRoutePattern(context.Background(), "GET /route")
	route, ok := requestctx.RoutePattern(ctx)
	if !ok || route != "GET /route" {
		t.Fatalf("expected route pattern to be set on background context")
	}
}
