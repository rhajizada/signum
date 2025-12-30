package requestctx

import (
	"context"
	"testing"
)

func TestEnsureAddsData(t *testing.T) {
	ctx := context.Background()
	if dataFrom(ctx) != nil {
		t.Fatalf("expected no data on base context")
	}

	ctx = Ensure(ctx)
	if dataFrom(ctx) == nil {
		t.Fatalf("expected data to be attached")
	}
}

func TestWithRoutePatternAndBackendID(t *testing.T) {
	ctx := context.Background()

	ctx = WithRoutePattern(ctx, "GET /api/badges/{id}")
	pattern, ok := RoutePattern(ctx)
	if !ok || pattern != "GET /api/badges/{id}" {
		t.Fatalf("expected route pattern to be set, got %q (ok=%v)", pattern, ok)
	}

	ctx = WithBackendID(ctx, "backend-1")
	backendID, ok := BackendID(ctx)
	if !ok || backendID != "backend-1" {
		t.Fatalf("expected backend id to be set, got %q (ok=%v)", backendID, ok)
	}
}

func TestWithBackendIDEmptyNoop(t *testing.T) {
	var ctx context.Context
	ctx = WithBackendID(ctx, "")
	if ctx != nil {
		t.Fatalf("expected nil context to remain nil")
	}
	if _, ok := BackendID(nil); ok {
		t.Fatalf("expected no backend id for nil context")
	}
}

func TestWithRoutePatternNilContext(t *testing.T) {
	ctx := WithRoutePattern(nil, "GET /route")
	route, ok := RoutePattern(ctx)
	if !ok || route != "GET /route" {
		t.Fatalf("expected route pattern to be set on background context")
	}
}
