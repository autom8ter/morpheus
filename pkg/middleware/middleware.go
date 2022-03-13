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

func (m *Middleware) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		bearToken := req.Header.Get("Authorization")
		token := strings.TrimPrefix(bearToken, "Bearer ")
		now := time.Now()
		w := &responseWriterWrapper{w: writer}
		logFields := map[string]interface{}{
			"url":    req.URL.String(),
			"method": req.Method,
		}
		defer func() {
			if w.statusCode == 0 {
				w.statusCode = 200
			}
			//logFields["response_body"] = w.Body()
			logFields["response_status"] = w.StatusCode()
			logFields["elapsed_ns"] = time.Since(now).Nanoseconds()
			logger.L.Info("http request/response", logFields)
		}()

		if token == "" {
			handler.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = context.WithValue(ctx, tokenCtxKey, token)
		claims, err := m.parseClaims(token)
		if err != nil {
			logger.L.Error("failed to parse Authorization token", map[string]interface{}{
				"error": err,
			})
		}
		if claims != nil {
			for _, usr := range m.config.Auth.Users {
				if usr.Username == claims["sub"] {
					logFields["user"] = usr.Username
					ctx = context.WithValue(ctx, userCtxKey, usr)
				}
			}
		}
		handler.ServeHTTP(w, req.WithContext(ctx))
	})
}
