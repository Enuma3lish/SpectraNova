package biz
package biz

import (
	"context"
	stderrs "errors"

	"MLW/fenzVideo/internal/pkg/hash"
	"MLW/fenzVideo/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

package biz

import (
	"context"
	stderrs "errors"

	"MLW/fenzVideo/internal/pkg/hash"
	"MLW/fenzVideo/internal/pkg/jwt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	DisplayName  string
	Role         string
	IsHidden     bool
}

type AuthRepo interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
}

type AuthUsecase struct {
	repo   AuthRepo
	tokens *jwt.Manager
	log    *log.Helper
}

func NewAuthUsecase(repo AuthRepo, tokens *jwt.Manager, logger log.Logger) *AuthUsecase {
	return &AuthUsecase{
		repo:   repo,
		tokens: tokens,
		log:    log.NewHelper(logger),
	}
}

var (
	ErrUserNotFound       = errors.NotFound("USER_NOT_FOUND", "user not found")
	ErrInvalidCredentials = errors.Unauthorized("INVALID_CREDENTIALS", "invalid credentials")
	ErrTokenInvalid       = errors.Unauthorized("TOKEN_INVALID", "token invalid")
	ErrUserHidden         = errors.Forbidden("USER_HIDDEN", "account is hidden")
	ErrUsernameTaken      = errors.Conflict("USERNAME_TAKEN", "username already exists")
)

func (uc *AuthUsecase) Register(ctx context.Context, username, password, displayName string) (*User, string, string, error) {
	_, err := uc.repo.FindByUsername(ctx, username)
	if err == nil {
		return nil, "", "", ErrUsernameTaken
	}
	if err != nil && !stderrs.Is(err, ErrUserNotFound) {
		return nil, "", "", err
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return nil, "", "", err
	}

	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Role:         "user",
	}

	created, err := uc.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	accessToken, refreshToken, err := uc.tokens.GenerateTokenPair(created.ID, created.Role)
	if err != nil {
		return nil, "", "", err
	}
	return created, accessToken, refreshToken, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*User, string, string, error) {
	user, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", "", err
	}
	if user.IsHidden {
		return nil, "", "", ErrUserHidden
	}
	if err := hash.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, refreshToken, err := uc.tokens.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		return nil, "", "", err
	}
	return user, accessToken, refreshToken, nil
}

func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := uc.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrTokenInvalid
	}
	accessToken, newRefreshToken, err := uc.tokens.GenerateTokenPair(claims.UserID, claims.Role)
	if err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}

func (uc *AuthUsecase) GetMe(ctx context.Context, userID int64) (*User, error) {
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.IsHidden {
		return nil, ErrUserHidden
	}
	return user, nil
}




}	return accessToken, newRefreshToken, nil	}		return "", "", err	if err != nil {	accessToken, newRefreshToken, err := uc.tokens.GenerateTokenPair(claims.UserID, claims.Role)	}		return "", "", ErrTokenInvalid	if err != nil {	claims, err := uc.tokens.ParseRefreshToken(refreshToken)func (uc *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (string, string, error) {}	return user, accessToken, refreshToken, nil	}		return nil, "", "", err	if err != nil {	accessToken, refreshToken, err := uc.tokens.GenerateTokenPair(user.ID, user.Role)	}		return nil, "", "", ErrInvalidCredentials	if err := hash.ComparePassword(user.PasswordHash, password); err != nil {	}		return nil, "", "", ErrUserHidden	if user.IsHidden {	}		return nil, "", "", err	if err != nil {	user, err := uc.repo.FindByUsername(ctx, username)func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (*User, string, string, error) {}	return created, accessToken, refreshToken, nil	}		return nil, "", "", err	if err != nil {	accessToken, refreshToken, err := uc.tokens.GenerateTokenPair(created.ID, created.Role)	}		return nil, "", "", err	if err != nil {	created, err := uc.repo.CreateUser(ctx, user)	}		Role:         "user",		DisplayName:  displayName,		PasswordHash: passwordHash,		Username:     username,	user := &User{	}		return nil, "", "", err	if err != nil {	passwordHash, err := hash.HashPassword(password)	}		return nil, "", "", err	if err != nil && !stderrs.Is(err, ErrUserNotFound) {	}		return nil, "", "", ErrUsernameTaken	if err == nil {	_, err := uc.repo.FindByUsername(ctx, username)func (uc *AuthUsecase) Register(ctx context.Context, username, password, displayName string) (*User, string, string, error) {)	ErrUsernameTaken      = errors.Conflict("USERNAME_TAKEN", "username already exists")