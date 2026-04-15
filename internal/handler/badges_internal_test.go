package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadgeHelpers(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "writes cache headers",
			run: func(t *testing.T) {
				rec := httptest.NewRecorder()
				writeBadgeCacheHeaders(rec, "etag-value", "time-value")
				headers := rec.Result().Header
				assert.NotEmpty(t, headers.Get("Cache-Control"))
				assert.Equal(t, "etag-value", headers.Get("ETag"))
				assert.Equal(t, "time-value", headers.Get("Last-Modified"))
			},
		},
		{
			name: "matches etag values",
			run: func(t *testing.T) {
				header := `W/"a", W/"b"`
				assert.True(t, etagMatches(header, `W/"b"`))
				assert.False(t, etagMatches(header, `W/"c"`))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.run)
	}
}
