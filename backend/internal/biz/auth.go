package biz

import (
	"context"
	"time"

	"backend/internal/conf"
	"backend/internal/pkg/hash"
	"backend/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type AuthUser struct {
	ID          uint64
	Username    string
	DisplayName string
	Password    string // hashed
	Role        string
	IsHidden    bool
	CreatedAt   time.Time
}

type AuthRepo interface {
	FindByUsername(ctx context.Context, username string) (*AuthUser, error)
	FindByID(ctx context.Context, id uint64) (*AuthUser, error)
	CreateUser(ctx context.Context, user *AuthUser) (*AuthUser, error)
	CreateChannel(ctx context.Context, userID uint64) error
}

type AuthUsecase struct {
	repo          AuthRepo
	jwtSecret     string
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
	log           *log.Helper
}

func NewAuthUsecase(repo AuthRepo, ac *conf.Auth, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		repo:          repo,
		jwtSecret:     ac.JwtSecret,
		tokenExpiry:   ac.TokenExpiry.AsDuration(),
		refreshExpiry: ac.RefreshExpiry.AsDuration(),
		log:           log.NewHelper(logger),
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, username, password, displayName string) (*AuthUser, string, string, error) {
	// Check username uniqueness
	existing, _ := uc.repo.FindByUsername(ctx, username)
	if existing != nil {
		return nil, "", "", errors.Conflict("USERNAME_ALREADY_EXISTS", "username already taken")
	}

	// Hash password
	hashed, err := hash.HashPassword(password)
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to hash password")
	}

	// Create user
	user, err := uc.repo.CreateUser(ctx, &AuthUser{
		Username:    username,
		DisplayName: displayName,
		Password:    hashed,
		Role:        "user",
	})
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to create user")
	}

	// Auto-create channel (every user is a potential creator)
	if err := uc.repo.CreateChannel(ctx, user.ID); err != nil {
		uc.log.Warnf("failed to create channel for user %d: %v", user.ID, err)
	}

	// Generate tokens
	token, err := jwt.GenerateToken(uc.jwtSecret, user.ID, user.Role, uc.tokenExpiry)
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to generate token")
	}
	refreshToken, err := jwt.GenerateRefreshToken(uc.jwtSecret, user.ID, uc.refreshExpiry)
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to generate refresh token")
	}

	return user, token, refreshToken, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*AuthUser, string, string, error) {
	user, err := uc.repo.FindByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, "", "", errors.Unauthorized("INVALID_CREDENTIALS", "invalid username or password")
	}

	if user.IsHidden {
		return nil, "", "", errors.Forbidden("ACCOUNT_HIDDEN", "account is hidden")
	}

	if !hash.ComparePassword(user.Password, password) {
		return nil, "", "", errors.Unauthorized("INVALID_CREDENTIALS", "invalid username or password")
	}

	token, err := jwt.GenerateToken(uc.jwtSecret, user.ID, user.Role, uc.tokenExpiry)
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to generate token")
	}
	refreshToken, err := jwt.GenerateRefreshToken(uc.jwtSecret, user.ID, uc.refreshExpiry)
	if err != nil {
		return nil, "", "", errors.InternalServer("INTERNAL", "failed to generate refresh token")
	}

	return user, token, refreshToken, nil
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	userID, err := jwt.ParseRefreshToken(uc.jwtSecret, refreshTokenStr)
	if err != nil {
		return "", "", errors.Unauthorized("TOKEN_INVALID", "invalid refresh token")
	}

	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return "", "", errors.NotFound("USER_NOT_FOUND", "user not found")
	}

	if user.IsHidden {
		return "", "", errors.Forbidden("ACCOUNT_HIDDEN", "account is hidden")
	}

	token, err := jwt.GenerateToken(uc.jwtSecret, user.ID, user.Role, uc.tokenExpiry)
	if err != nil {
		return "", "", errors.InternalServer("INTERNAL", "failed to generate token")
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(uc.jwtSecret, user.ID, uc.refreshExpiry)
	if err != nil {
		return "", "", errors.InternalServer("INTERNAL", "failed to generate refresh token")
	}

	return token, newRefreshToken, nil
}
