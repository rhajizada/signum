package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rhajizada/signum/internal/service"
)

func TestHomeHandler(t *testing.T) {
	repo := &fakeRepo{}
	tokens, err := service.NewTokenManager("secret")
	if err != nil {
		t.Fatalf("token manager: %v", err)
	}
	h := newHandler(t, repo, tokens)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	h.Home(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status ok, got %d", rec.Code)
	}
	if got := rec.Result().Header.Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("expected html content type, got %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `value="build"`) || !strings.Contains(body, `value="passing"`) {
		t.Fatalf("expected template values in body")
	}
}
