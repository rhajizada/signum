package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rhajizada/signum/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHomeHandler(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{name: "renders home template", host: "example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokens, err := service.NewTokenManager("secret")
			require.NoError(t, err)
			h := newHandler(t, &fakeRepo{}, tokens)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tc.host
			rec := httptest.NewRecorder()
			h.Home(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Result().Header.Get("Content-Type"), "text/html")
			assert.Contains(t, rec.Body.String(), `value="build"`)
			assert.Contains(t, rec.Body.String(), `value="passing"`)
		})
	}
}
