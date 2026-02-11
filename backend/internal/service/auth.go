package service

import (
	"context"

	"MLW/fenzVideo/internal/biz"
	"MLW/fenzVideo/internal/pkg/authctx"

	"github.com/go-kratos/kratos/v2/log"

	v1 "MLW/fenzVideo/api/fenzvideo/v1"
)

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	uc  *biz.AuthUsecase
	log *log.Helper
}

func NewAuthService(uc *biz.AuthUsecase, logger log.Logger) *AuthService {
	return &AuthService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	user, accessToken, refreshToken, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		UserId:       user.ID,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	user, accessToken, refreshToken, err := s.uc.Register(ctx, req.Username, req.Password, req.DisplayName)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		UserId:       user.ID,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenReply, error) {
	accessToken, refreshToken, err := s.uc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) GetMe(ctx context.Context, _ *v1.GetMeRequest) (*v1.GetMeReply, error) {
	userID, ok := authctx.CurrentUserID(ctx)
	if !ok {
		return nil, authctx.ErrUnauthorized
	}
	user, err := s.uc.GetMe(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &v1.GetMeReply{
		UserId:      user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	}, nil
}
