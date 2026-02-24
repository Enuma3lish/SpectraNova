package authctx

import "context"

type contextKey string

const (
	ContextKeyUserID contextKey = "user_id"
	ContextKeyRole   contextKey = "role"
)

// WithUserID sets user_id in context.
func WithUserID(ctx context.Context, uid uint64) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, uid)
}

// WithRole sets role in context.
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ContextKeyRole, role)
}

// UserIDFromContext extracts user_id from context.
func UserIDFromContext(ctx context.Context) (uint64, bool) {
	uid, ok := ctx.Value(ContextKeyUserID).(uint64)
	return uid, ok
}

// RoleFromContext extracts role from context.
func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ContextKeyRole).(string)
	return role, ok
}
