package auth

import (
	"errors"
	"net/http"
	"sync"
)

type AuthFunc func(w http.ResponseWriter, r *http.Request) error

func Middleware(auth AuthFunc, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := auth(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func BasicAuth(users map[string]string) AuthFunc {
	mu := sync.RWMutex{}
	return func(w http.ResponseWriter, r *http.Request) error {
		usr, pw, ok := r.BasicAuth()
		if !ok || usr == "" {
			return errors.New("basicauth: missing username/password")
		}
		mu.RLock()
		defer mu.RUnlock()
		if pw != users[usr] {
			return errors.New("basicauth: incorrect password")
		}
		return nil
	}
}
