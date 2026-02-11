package authctx
package authctx

import (
	"context"




























}	return role, ok	role, ok := ctx.Value(ctxRoleKey).(string)func CurrentRole(ctx context.Context) (string, bool) {}	return userID, ok	userID, ok := ctx.Value(ctxUserIDKey).(int64)func CurrentUserID(ctx context.Context) (int64, bool) {}	return ctx	ctx = context.WithValue(ctx, ctxRoleKey, role)	ctx = context.WithValue(ctx, ctxUserIDKey, userID)func WithUser(ctx context.Context, userID int64, role string) context.Context {var ErrUnauthorized = errors.Unauthorized("TOKEN_INVALID", "unauthorized"))	ctxRoleKey   ctxKey = "role"	ctxUserIDKey ctxKey = "user_id"const (type ctxKey string)	"github.com/go-kratos/kratos/v2/errors"