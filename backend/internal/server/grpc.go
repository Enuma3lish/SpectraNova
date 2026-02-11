package server

import (
	"MLW/fenzVideo/internal/conf"
	"MLW/fenzVideo/internal/pkg/jwt"
	"MLW/fenzVideo/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "MLW/fenzVideo/api/fenzvideo/v1"
)

func NewGRPCServer(c conf.Server, auth *service.AuthService, tokens *jwt.Manager, logger log.Logger) *grpc.Server {
	options := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			AuthnMiddleware(tokens),
			selector.Server(AuthRequired).Build(RequireAuth()),
		),
	}
	if c.GRPC.Addr != "" {
		options = append(options, grpc.Address(c.GRPC.Addr))
	}
	if c.GRPC.Timeout > 0 {
		options = append(options, grpc.Timeout(c.GRPC.Timeout))
	}

	srv := grpc.NewServer(options...)
	v1.RegisterAuthServiceServer(srv, auth)
	return srv
}
