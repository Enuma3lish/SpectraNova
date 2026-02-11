package server

import (
	"context"
	"strings"

	"MLW/fenzVideo/internal/pkg/authctx"
	"MLW/fenzVideo/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func AuthnMiddleware(tokens *jwt.Manager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if ok {
				token := bearerToken(tr.RequestHeader().Get("Authorization"))
				if token != "" {
					claims, err := tokens.ParseAccessToken(token)
					if err != nil {
						return nil, authctx.ErrUnauthorized
					}
					ctx = authctx.WithUser(ctx, claims.UserID, claims.Role)
				}
			}
			return handler(ctx, req)
		}
	}
}

func RequireAuth() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if _, ok := authctx.CurrentUserID(ctx); !ok {
				return nil, authctx.ErrUnauthorized
			}
			return handler(ctx, req)
		}
	}
}

func AuthRequired(ctx context.Context, operation string) bool {
	return strings.HasSuffix(operation, "AuthService/GetMe")
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
