package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
)

type AuthService struct {
	v1.UnimplementedAuthServiceServer
	uc *biz.AuthUsecase
}

func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.RegisterReply, error) {
	user, token, refreshToken, err := s.uc.Register(ctx, req.Username, req.Password, req.DisplayName)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterReply{
		Id:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	user, token, refreshToken, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Id:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Role:         user.Role,
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.RefreshTokenRequest) (*v1.RefreshTokenReply, error) {
	token, refreshToken, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &v1.RefreshTokenReply{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}
