package middleware

import (
	"context"
	"fmt"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/autom8ter/morpheus/pkg/version"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/palantir/stacktrace"
	"time"
)

const (
	userCtxKey  = "auth_user_ctx_key"
	tokenCtxKey = "auth_token_ctx_key"
)

func GetUserCtx(ctx context.Context) (config.User, bool) {
	val, ok := ctx.Value(userCtxKey).(config.User)
	return val, ok
}

func GetTokenCtx(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(tokenCtxKey).(string)
	return val, ok
}

func (a *Middleware) RequireRole(ctx context.Context, role config.Role) (config.User, error) {
	usr, ok := GetUserCtx(ctx)
	if !ok {
		return config.User{}, stacktrace.Propagate(constants.ErrUnauthorized, "no context user")
	}
	for _, usrRole := range usr.Roles {
		if usrRole == role || usrRole == config.ADMIN {
			return usr, nil
		}
	}
	return config.User{}, stacktrace.Propagate(constants.ErrForbidden, "user %s missing role %s", usr.Username, role)
}

func (a *Middleware) RequireAuth(ctx context.Context) (config.User, error) {
	usr, ok := GetUserCtx(ctx)
	if !ok {
		return config.User{}, stacktrace.Propagate(constants.ErrUnauthorized, "no user in context")
	}
	return usr, nil
}

func (a *Middleware) parseClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check jwt signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, stacktrace.NewError("Unexpected signing method: %v", token.Header["alg"])
		}
		// return secret
		return []byte(a.config.Auth.SigningSecret), nil
	})
	if err != nil {
		return jwt.MapClaims{}, stacktrace.Propagate(err, "failed to parse token")
	}
	if !token.Valid {
		return jwt.MapClaims{}, stacktrace.NewError("invalid token")
	}
	return token.Claims.(jwt.MapClaims), nil
}

func (a *Middleware) Login(username, password string) (string, error) {
	var matched = false
	for _, usr := range a.config.Auth.Users {
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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Audience:  fmt.Sprintf("%s-%s", constants.JWTAudience, version.Version),
		ExpiresAt: now.Add(a.config.Auth.TokenTTL).Unix(),
		Id:        uuid.NewString(),
		IssuedAt:  now.Unix(),
		Issuer:    fmt.Sprintf("%s-%s", constants.JWTAudience, version.Version),
		NotBefore: now.Unix() - skew,
		Subject:   username,
	})
	tokenString, err := token.SignedString([]byte(a.config.Auth.SigningSecret))
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	return tokenString, nil
}
