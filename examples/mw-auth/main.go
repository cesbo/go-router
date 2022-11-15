package main

import (
	"context"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/cesbo/go-router"
)

type userContext struct {
	Name string
}

func (userContext) String() string {
	return "userContext"
}

// Key for request Context
var userContextKey = &userContext{}

var loginForm = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Login</title>
</head>
<body>
	<form method="POST" action="/login">
		<input type="text" name="name" placeholder="Name" />
		<input type="submit" value="Login" />
	</form>
</body>
</html>`))

var helloForm = template.Must(template.New("hello").Parse(`<!DOCTYPE html>
<html>
<head>
	<title>Hello</title>
</head>
<body>
	<form method="POST" action="/logout">
		<b>Hello, {{ .Name }}!</b>
		<input type="submit" value="Logout" />
	</form>
</body>
</html>`))

// AuthMW is a middleware that checks for a cookie named "auth"
// If cookie found, it adds userContext to the request context
func AuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err == nil {
			r = r.WithContext(
				context.WithValue(
					r.Context(),
					userContextKey,
					&userContext{
						Name: cookie.Value,
					},
				),
			)
		}
		next.ServeHTTP(w, r)
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(*userContext)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		loginForm.Execute(w, nil)
	} else {
		if err := helloForm.Execute(w, user); err != nil {
			log.Println(err)
		}
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:    "auth",
			Path:    "/",
			Value:   r.PostForm.Get("name"),
			Expires: time.Now().Add(5 * time.Minute),
		},
	)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:   "auth",
			Path:   "/",
			Value:  "",
			MaxAge: -1,
		},
	)
	http.Redirect(w, r, "/", http.StatusFound)
}

func main() {
	r := router.NewRouter()

	// Add middleware
	r.Use(AuthMW)

	// Add routes
	r.HandleFunc("/", index)
	r.HandleFunc("/login", login)
	r.HandleFunc("/logout", logout)

	// Start server
	http.ListenAndServe(":8080", r)
}
