package router

import (
	"net/http"
	"sync"
)

type MiddlewareFunc func(http.Handler) http.Handler

// Router is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that most closely matches
// the URL. There are two types of patterns:
// 1. Exact match: /foo/bar
// 2. Wildcard match: /foo/
type Router struct {
	mutex sync.RWMutex
	radix *Radix
	mw    []MiddlewareFunc

	NotFoundHandler http.Handler
}

// NewRouter returns a new router.
func NewRouter() *Router {
	return &Router{
		radix:           new(Radix),
		NotFoundHandler: http.NotFoundHandler(),
	}
}

func (r *Router) Use(mw ...MiddlewareFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.mw = append(r.mw, mw...)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle replaces it.
func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.radix.Insert(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
// If a handler already exists for pattern, HandleFunc replaces it.
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, handler)
}

// Lookup returns the handler for the given path.
// If no handler is found, it returns the NotFound handler.
func (r *Router) Lookup(path string) http.Handler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if handler, ok := r.radix.LookupPath(path).(http.Handler); ok {
		return handler
	}

	return nil
}

// Remove removes the handler for the given path.
func (r *Router) Remove(path string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.radix.Remove(path)
}

func (r *Router) prepare(path string) http.Handler {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if handler, ok := r.radix.LookupPath(path).(http.Handler); ok {
		for i := len(r.mw) - 1; i >= 0; i-- {
			handler = r.mw[i](handler)
		}

		return handler
	}

	return r.NotFoundHandler
}

// ServeHTTP dispatches the request to the handler whose pattern most closely
// matches the request URL.
func (r *Router) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	r.prepare(request.URL.Path).ServeHTTP(response, request)
}
