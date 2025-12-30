package handler

import (
	"net/http/httptest"
	"testing"
)

func TestWriteBadgeCacheHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	writeBadgeCacheHeaders(rec, "etag-value", "time-value")
	headers := rec.Result().Header

	if got := headers.Get("Cache-Control"); got == "" {
		t.Fatalf("expected cache-control header")
	}
	if got := headers.Get("ETag"); got != "etag-value" {
		t.Fatalf("expected etag header, got %q", got)
	}
	if got := headers.Get("Last-Modified"); got != "time-value" {
		t.Fatalf("expected last-modified header, got %q", got)
	}
}

func TestEtagMatches(t *testing.T) {
	header := `W/"a", W/"b"`
	if !etagMatches(header, `W/"b"`) {
		t.Fatalf("expected etag match")
	}
	if etagMatches(header, `W/"c"`) {
		t.Fatalf("expected etag mismatch")
	}
}
