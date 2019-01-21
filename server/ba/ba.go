package ba

import (
	"fmt"
	"net/http"
)

var (
	username string
	password string
)

func SetUserPassword(user, pass string) {
	username = user
	password = pass
}

var (
	AuthFunc func(string, string) bool
)

func cb(u, p string) bool {
	if AuthFunc == nil {
		return false
	}
	return AuthFunc(u, p)
}

func HandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if user != username || pass != password || !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please login"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n"))
			return
		}
		next.ServeHTTP(w, r)
	}
}

func HandlerFuncCB(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !cb(user, pass) || !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please login"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n"))
			return
		}
		next.ServeHTTP(w, r)
	}
}

func Handler(next http.Handler) http.Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func DisallowRobots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "User-agent: *")
	fmt.Fprintln(w, "Disallow: /")
}
