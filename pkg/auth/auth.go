package auth

import (
	"context"
	"errors"
	"github.com/autom8ter/morpheus/pkg/config"
	"net/http"
	"sync"
)

type AuthFunc func(w http.ResponseWriter, r *http.Request) (context.Context, error)

func Middleware(auth AuthFunc, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := auth(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

const userKey = "morpheus-user"

func BasicAuth(users []*config.User) AuthFunc {
	mu := sync.RWMutex{}
	return func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		usr, pw, ok := r.BasicAuth()
		if !ok || usr == "" {
			return nil, errors.New("basicauth: missing username/password")
		}
		mu.RLock()
		defer mu.RUnlock()
		for _, user := range users {
			if user.Username == usr && user.Password == pw {
				return context.WithValue(r.Context(), userKey, user), nil
			}
		}
		return nil, errors.New("basicauth: authentication failed")
	}
}

func GetUser(ctx context.Context) (*config.User, bool) {
	if ctx.Value(userKey) == nil {
		return nil, false
	}
	val, ok := ctx.Value(userKey).(*config.User)
	return val, ok
}
