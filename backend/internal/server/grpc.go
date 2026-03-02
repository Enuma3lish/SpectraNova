package server

import (
	v1 "backend/api/fenzvideo/v1"
	"backend/internal/conf"
	"backend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	ac *conf.Auth,
	logger log.Logger,
	authSvc *service.AuthService,
	categorySvc *service.CategoryService,
	tagSvc *service.TagService,
	videoSvc *service.VideoService,
	searchSvc *service.SearchService,
	channelSvc *service.ChannelService,
	adminSvc *service.AdminService,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			JWTAuthMiddleware(ac.JwtSecret),
			AdminGuardMiddleware(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)

	// Register all gRPC services
	v1.RegisterAuthServiceServer(srv, authSvc)
	v1.RegisterCategoryServiceServer(srv, categorySvc)
	v1.RegisterTagServiceServer(srv, tagSvc)
	v1.RegisterVideoServiceServer(srv, videoSvc)
	v1.RegisterSearchServiceServer(srv, searchSvc)
	v1.RegisterChannelServiceServer(srv, channelSvc)
	v1.RegisterAdminServiceServer(srv, adminSvc)

	return srv
}
