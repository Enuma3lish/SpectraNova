package server

import (
	"context"
	"strings"

	"backend/internal/pkg/authctx"
	pjwt "backend/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// publicPaths are Kratos operation names that do not require authentication.
// Kratos sets operation via protoc-gen-go-http: "/package.ServiceName/MethodName"
var publicPaths = map[string]bool{
	"/fenzvideo.v1.AuthService/Login":        true,
	"/fenzvideo.v1.AuthService/Register":     true,
	"/fenzvideo.v1.AuthService/RefreshToken": true,
}

// publicPrefixes are operation name prefixes that do not require authentication.
var publicPrefixes = []string{
	"/fenzvideo.v1.VideoService/GetRecommended",
	"/fenzvideo.v1.VideoService/GetVideo",
	"/fenzvideo.v1.SearchService/",
	"/fenzvideo.v1.CategoryService/",
	"/fenzvideo.v1.ChannelService/GetChannel",
	"/fenzvideo.v1.TagService/", // all tag ops public (guest session_id support)
}

func JWTAuthMiddleware(jwtSecret string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			operation := tr.Operation()

			// Public paths: attempt optional token extraction but don't require it
			if isPublicPath(operation) {
				authHeader := tr.RequestHeader().Get("Authorization")
				if authHeader != "" {
					tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
					if tokenStr != authHeader {
						claims, err := pjwt.ParseToken(jwtSecret, tokenStr)
						if err == nil {
							ctx = authctx.WithUserID(ctx, claims.UserID)
							ctx = authctx.WithRole(ctx, claims.Role)
						}
					}
				}
				return handler(ctx, req)
			}

			// Protected paths: require token
			authHeader := tr.RequestHeader().Get("Authorization")
			if authHeader == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "missing authorization header")
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				return nil, errors.Unauthorized("UNAUTHORIZED", "invalid authorization format")
			}

			claims, err := pjwt.ParseToken(jwtSecret, tokenStr)
			if err != nil {
				return nil, errors.Unauthorized("TOKEN_INVALID", "invalid or expired token")
			}

			ctx = authctx.WithUserID(ctx, claims.UserID)
			ctx = authctx.WithRole(ctx, claims.Role)

			return handler(ctx, req)
		}
	}
}

func AdminGuardMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			operation := tr.Operation()
			if !strings.Contains(operation, "Admin") {
				return handler(ctx, req)
			}

			role, ok := authctx.RoleFromContext(ctx)
			if !ok || role != "admin" {
				return nil, errors.Forbidden("ADMIN_REQUIRED", "admin role required")
			}

			return handler(ctx, req)
		}
	}
}

func isPublicPath(operation string) bool {
	if publicPaths[operation] {
		return true
	}
	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(operation, prefix) {
			return true
		}
	}
	return false
}
