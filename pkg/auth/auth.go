package auth

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/logger"
	"github.com/autom8ter/morpheus/pkg/version"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"net/http"
	"strings"
	"time"
)

type Auth struct {
	config *config.Auth
}

func NewAuth(config *config.Auth) *Auth {
	return &Auth{config: config}
}

func (a *Auth) parseClaims(tokenString string) (Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check jwt signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, stacktrace.NewError("Unexpected signing method: %v", token.Header["alg"])
		}
		// return secret
		return a.config.SigningSecret, nil
	})

	if err != nil {
		return Claims{}, stacktrace.Propagate(err, "failed to parse token")
	}
	if !token.Valid {
		return Claims{}, stacktrace.NewError("invalid token")
	}
	return token.Claims.(Claims), nil
}

func (a *Auth) Login(username, password string) (string, error) {
	var matched = false
	for _, usr := range a.config.Users {
		if usr.Username == username && usr.Password == password {
			matched = true
			break
		}
	}
	if !matched {
		return "", constants.ErrUnauthorized
	}
	skew := int64(500)
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  fmt.Sprintf("%s-%s", constants.JWTAudience, version.Version),
			ExpiresAt: now.Add(a.config.TokenTTL).Unix(),
			Id:        uuid.NewString(),
			IssuedAt:  now.Unix(),
			Issuer:    fmt.Sprintf("%s-%s", constants.JWTAudience, version.Version),
			NotBefore: now.Unix() - skew,
			Subject:   username,
		},
	})
	tokenString, err := token.SignedString(a.config.SigningSecret)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	return tokenString, nil
}

func (a *Auth) JwtClaimsParser(handler http.Handler, require bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
		if token == "" && !require {
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if token != "" {
			ctx = context.WithValue(r.Context(), tokenCtxKey, token)
		}
		claims, err := a.parseClaims(token)
		if err != nil && require {
			logger.L.HTTPError(w, "failed to parse Authorization token", err, http.StatusUnauthorized)
			return
		}

		for _, usr := range a.config.Users {
			if usr.Username == claims.Subject {
				handler.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userCtxKey, usr)))
				return
			}
		}
		if require {
			logger.L.HTTPError(w, "failed to authenticate", err, http.StatusUnauthorized)
			return
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
