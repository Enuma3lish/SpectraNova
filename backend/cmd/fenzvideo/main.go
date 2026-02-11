package main

import (
	"flag"
	"os"

	"MLW/fenzVideo/internal/biz"
	"MLW/fenzVideo/internal/conf"
	"MLW/fenzVideo/internal/data"
	"MLW/fenzVideo/internal/pkg/jwt"
	"MLW/fenzVideo/internal/server"
	"MLW/fenzVideo/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	confPath = flag.String("conf", "configs/config.yaml", "config path")
)

func main() {
	flag.Parse()

	logger := log.NewStdLogger(os.Stdout)
	c := config.New(
		config.WithSource(
			file.NewSource(*confPath),
		),
	)
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bootstrap conf.Bootstrap
	if err := c.Scan(&bootstrap); err != nil {
		panic(err)
	}

	dataLayer, err := data.NewData(bootstrap.Data, logger)
	if err != nil {
		panic(err)
	}
	authRepo := data.NewAuthRepo(dataLayer, logger)
	tokenManager := jwt.NewManager(bootstrap.Auth.JWTSecret, bootstrap.Auth.AccessExpiry, bootstrap.Auth.RefreshExpiry)
	usecase := biz.NewAuthUsecase(authRepo, tokenManager, logger)
	authService := service.NewAuthService(usecase, logger)

	httpSrv := server.NewHTTPServer(bootstrap.Server, authService, tokenManager, logger)
	grpcSrv := server.NewGRPCServer(bootstrap.Server, authService, tokenManager, logger)

	app := kratos.New(
		kratos.Name("fenzvideo"),
		kratos.Logger(logger),
		kratos.Server(
			httpSrv,
			grpcSrv,
		),
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
