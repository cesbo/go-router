package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testHandler int

func (testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func TestRouter_NewRouter(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	// not found by default
	assert.HTTPStatusCode(
		router.ServeHTTP,
		http.MethodGet,
		"/not/found",
		nil,
		http.StatusNotFound,
	)

	// redefine root handler
	rootHandler := testHandler(1)
	router.Handle("/", rootHandler)

	// root handler is used
	assert.HTTPStatusCode(
		router.ServeHTTP,
		http.MethodGet,
		"/found",
		nil,
		http.StatusNoContent,
	)
}

func TestRouter_Handle(t *testing.T) {
	router := NewRouter()

	rootHandler := testHandler(1)
	router.Handle("/", rootHandler)

	t.Run("entry without root", func(t *testing.T) {
		assert := assert.New(t)

		fooHandler := testHandler(11)
		router.Handle("/a/b/c/foo", fooHandler)
		barHandler := testHandler(12)
		router.Handle("/a/b/c/bar", barHandler)
		bazHandler := testHandler(13)
		router.Handle("/a/b/c/baz", bazHandler)

		if r := router.Lookup("/a/b/c"); !assert.Exactly(rootHandler, r) {
			return
		}
		if r := router.Lookup("/a/b/c/foo"); !assert.Exactly(fooHandler, r) {
			return
		}
		if r := router.Lookup("/a/b/c/bar"); !assert.Exactly(barHandler, r) {
			return
		}
		if r := router.Lookup("/a/b/c/baz"); !assert.Exactly(bazHandler, r) {
			return
		}
		if r := router.Lookup("/a/b/c/etc"); !assert.Exactly(rootHandler, r) {
			return
		}
	})

	t.Run("with root", func(t *testing.T) {
		assert := assert.New(t)

		nestedItemHandler := testHandler(21)
		router.Handle("/d/e/f", nestedItemHandler)
		nestedRootHandler := testHandler(22)
		router.Handle("/d/e/f/", nestedRootHandler)
		fooHandler := testHandler(23)
		router.Handle("/d/e/f/foo", fooHandler)

		if r := router.Lookup("/d/e/f"); !assert.Exactly(nestedItemHandler, r) {
			return
		}
		if r := router.Lookup("/d/e/f/foo"); !assert.Exactly(fooHandler, r) {
			return
		}
		if r := router.Lookup("/d/e/f/etc"); !assert.Exactly(nestedRootHandler, r) {
			return
		}
	})
}

func TestRouter_Lookup(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	h1 := testHandler(1)
	router.Handle("/", h1)
	if r := router.Lookup("/"); !assert.Exactly(h1, r) {
		return
	}
	if r := router.Lookup("/a/b/c"); !assert.Exactly(h1, r) {
		return
	}

	// /a is a item handler. full address should point to nearest root handler
	h2 := testHandler(2)
	router.Handle("/a", h2)
	if r := router.Lookup("/a"); !assert.Exactly(h2, r) {
		return
	}
	if r := router.Lookup("/a/b/c"); !assert.Exactly(h1, r) {
		return
	}

	// /a/b is a item handler. same as above
	h3 := testHandler(3)
	router.Handle("/a/b", h3)

	if r := router.Lookup("/a/b"); !assert.Exactly(h3, r) {
		return
	}
	if r := router.Lookup("/a/b/c"); !assert.Exactly(h1, r) {
		return
	}

	// /a/ is a root handler. full address should point to it
	h4 := testHandler(4)
	router.Handle("/a/", h4)
	if r := router.Lookup("/a/b/"); !assert.Exactly(h4, r) {
		return
	}
	if r := router.Lookup("/a/b/c"); !assert.Exactly(h4, r) {
		return
	}

	// entry without handlers should point to nearest root handler (h4)
	h5 := testHandler(5)
	router.Handle("/a/b/c/d/e/f", h5)
	if r := router.Lookup("/a/b/c"); !assert.Exactly(h4, r) {
		return
	}
}

func TestRouter_Use(t *testing.T) {
	assert := assert.New(t)
	router := NewRouter()

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "1")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "2")
			next.ServeHTTP(w, r)
		})
	}

	router.Use(mw1, mw2)

	handler := testHandler(1)
	router.Handle("/", handler)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, r)

	response := w.Result()
	assert.Equal(http.StatusNoContent, response.StatusCode)
	assert.Equal("1", response.Header.Get("X-Middleware-1"))
	assert.Equal("2", response.Header.Get("X-Middleware-2"))
}
