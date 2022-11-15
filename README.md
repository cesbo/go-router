# go-router

[![Go Reference](https://pkg.go.dev/badge/github.com/cesbo/go-router.svg)](https://pkg.go.dev/github.com/cesbo/go-router)

HTTP Router

Features:

- Compatible with net/http
- Fast search with radix tree
- Extremely lightweight:
    - No path variables (use query variables, they are good serialized)
    - No methods routing
    - No regexp
- Middleware support only on the router level

## Installation

To install the library use the following command in the project directory:

```
go get github.com/cesbo/go-router
```

## Quick Start

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    if !query.Has("name") {
        fmt.Fprintf(w, "Usage: /hello?name=John")
    } else {
        name := query.Get("name")
        fmt.Fprintf(w, "Hello, %s!", name)
    }
}

func wildcardHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Resource path: " + r.URL.Path))
}

r := router.NewRouter()

r.HandleFunc("/hello", helloHandler)
r.HandleFunc("/static/", wildcardHandler)

http.ListenAndServe(":8080", r)
```

## Handlers

Handler path could be two types:

- Resource path with exact path match. Example: `/hello`
- Directory path with prefix match. Example: `/static/`

## Middleware

Middleware is a function that is called before the handler.
For example you can use it to log request information:

```go
func LogMW(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello, World!"))
}

r := router.NewRouter()

r.Use(LogMW)
r.HandleFunc("/hello", helloHandler)

http.ListenAndServe(":8080", r)
```
