package middleware

import (
	"context"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/logger"
	"net/http"
	"strings"
	"time"
)

type Middleware struct {
	config *config.Config
}

func NewMiddleware(config *config.Config) *Middleware {
	return &Middleware{config: config}
}

func (m *Middleware) Wrap(handler http.Handler, require bool) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		bearToken := req.Header.Get("Authorization")
		token := strings.TrimPrefix(bearToken, "Bearer ")
		now := time.Now()
		w := &responseWriterWrapper{w: writer}
		defer func() {
			body := w.Body()
			fields := map[string]interface{}{
				"url":             req.URL.String(),
				"method":          req.Method,
				"elapsed_ns":      time.Since(now).Nanoseconds(),
				"response_status": w.StatusCode(),
				"response_body":   body,
			}
			logger.L.Info("http request/response", fields)
		}()

		if token == "" && !require {
			handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		if token != "" {
			ctx = context.WithValue(ctx, tokenCtxKey, token)
		}
		claims, err := m.parseClaims(token)
		if err != nil {
			logger.L.HTTPError(w, "failed to parse Authorization token", err, http.StatusUnauthorized)
			return
		}

		for _, usr := range m.config.Auth.Users {
			if usr.Username == claims["sub"] {
				ctx = context.WithValue(ctx, userCtxKey, usr)
				handler.ServeHTTP(w, req.WithContext(ctx))
				return
			}
		}
		if require {
			logger.L.HTTPError(w, "failed to authenticate", err, http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, req)
	})
}
