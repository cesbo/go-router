package main

import (
	"log"
	"net/http"

	"github.com/cesbo/go-router"
)

func LogMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := router.NewRouter()

	r.Use(LogMW)
	r.HandleFunc("/", rootHandler)

	http.ListenAndServe(":8080", r)
}
