package auth

import (
	"context"
	"github.com/autom8ter/morpheus/pkg/config"
	"github.com/autom8ter/morpheus/pkg/constants"
	"github.com/palantir/stacktrace"
)

const (
	userCtxKey  = "auth_user_ctx_key"
	tokenCtxKey = "auth_token_ctx_key"
)

func (a *Auth) getUserCtx(ctx context.Context) (config.User, bool) {
	val, ok := ctx.Value(userCtxKey).(config.User)
	return val, ok
}

func (a *Auth) getTokenCtx(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(tokenCtxKey).(string)
	return val, ok
}

func (a *Auth) RequireRole(ctx context.Context, role config.Role) (config.User, error) {
	usr, ok := a.getUserCtx(ctx)
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

func (a *Auth) RequireAuth(ctx context.Context) (config.User, error) {
	usr, ok := a.getUserCtx(ctx)
	if !ok {
		return config.User{}, stacktrace.Propagate(constants.ErrUnauthorized, "no user in context")
	}
	return usr, nil
}
