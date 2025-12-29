package router

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/rhajizada/signum/internal/handler"
	"github.com/rhajizada/signum/internal/requestctx"
)

// Router wires URL paths to handler methods.
type Router struct {
	mux *http.ServeMux
}

// New builds the HTTP routing table.
func New(h *handler.Handler) *Router {
	r := &Router{
		mux: http.NewServeMux(),
	}
	r.mux.Handle("/api/docs/", httpSwagger.WrapHandler)
	r.Handle("GET /api/live", http.HandlerFunc(h.LiveBadge))
	return r
}

// Handle registers a route with optional middleware wrappers.
func (r *Router) Handle(path string, handler http.Handler, wrappers ...func(http.Handler) http.Handler) {
	base := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := requestctx.WithRoutePattern(req.Context(), path)
		handler.ServeHTTP(w, req.WithContext(ctx))
	})

	var wrapped http.Handler = base
	for i := len(wrappers) - 1; i >= 0; i-- {
		wrapped = wrappers[i](wrapped)
	}
	r.mux.Handle(path, wrapped)
}
