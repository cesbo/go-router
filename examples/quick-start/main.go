package main

import (
	"fmt"
	"net/http"

	"github.com/cesbo/go-router"
)

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

func main() {
	r := router.NewRouter()

	r.HandleFunc("/hello", helloHandler)
	r.HandleFunc("/static/", wildcardHandler)

	http.ListenAndServe(":8080", r)
}
