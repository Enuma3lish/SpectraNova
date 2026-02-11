package server

import (
	"MLW/fenzVideo/internal/conf"
	"MLW/fenzVideo/internal/pkg/jwt"
	"MLW/fenzVideo/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"

	v1 "MLW/fenzVideo/api/fenzvideo/v1"
)

func NewHTTPServer(c conf.Server, auth *service.AuthService, tokens *jwt.Manager, logger log.Logger) *http.Server {
	options := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			AuthnMiddleware(tokens),
			selector.Server(AuthRequired).Build(RequireAuth()),
		),
	}
	if c.HTTP.Addr != "" {
		options = append(options, http.Address(c.HTTP.Addr))
	}
	if c.HTTP.Timeout > 0 {
		options = append(options, http.Timeout(c.HTTP.Timeout))
	}

	srv := http.NewServer(options...)
	v1.RegisterAuthServiceHTTPServer(srv, auth)
	return srv
}
