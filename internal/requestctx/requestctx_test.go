package requestctx_test

import (
	"context"
	"testing"

	"github.com/rhajizada/signum/internal/requestctx"
	"github.com/stretchr/testify/assert"
)

func TestContextHelpers(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "ensure adds request data container",
			run: func(t *testing.T) {
				ctx := requestctx.Ensure(context.Background())
				route, ok := requestctx.RoutePattern(ctx)
				assert.False(t, ok)
				assert.Empty(t, route)
			},
		},
		{
			name: "stores route pattern and backend id",
			run: func(t *testing.T) {
				ctx := requestctx.WithRoutePattern(context.Background(), "GET /api/badges/{id}")
				pattern, ok := requestctx.RoutePattern(ctx)
				assert.True(t, ok)
				assert.Equal(t, "GET /api/badges/{id}", pattern)

				ctx = requestctx.WithBackendID(ctx, "backend-1")
				backendID, ok := requestctx.BackendID(ctx)
				assert.True(t, ok)
				assert.Equal(t, "backend-1", backendID)
			},
		},
		{
			name: "empty backend id is noop for nil context",
			run: func(t *testing.T) {
				var ctx context.Context
				ctx = requestctx.WithBackendID(ctx, "")
				assert.Nil(t, ctx)
				_, ok := requestctx.BackendID(context.Background())
				assert.False(t, ok)
			},
		},
		{
			name: "writes route pattern on background context",
			run: func(t *testing.T) {
				ctx := requestctx.WithRoutePattern(context.Background(), "GET /route")
				route, ok := requestctx.RoutePattern(ctx)
				assert.True(t, ok)
				assert.Equal(t, "GET /route", route)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}
