package router

import (
	"net/http"
)

// Middleware type
type Middleware func(http.HandlerFunc) http.HandlerFunc

type Router struct {
	prefix      string
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router) Group(prefix string) *Router {
	mws := make([]Middleware, len(r.middlewares))
	copy(mws, r.middlewares)
	return &Router{
		prefix:      r.prefix + prefix,
		mux:         r.mux,
		middlewares: mws,
	}
}

// Use adds middleware(s) to the router
func (r *Router) Use(mw ...Middleware) {
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) HandleFunc(path string, handler http.HandlerFunc) {
	fullPath := r.prefix + path

	// Apply middlewares in reverse order (outermost first)
	h := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	r.mux.HandleFunc(fullPath, h)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
