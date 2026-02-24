package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type authRepo struct {
	data *Data
	log  *log.Helper
}

func NewAuthRepo(data *Data, logger log.Logger) biz.AuthRepo {
	return &authRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *authRepo) FindByUsername(ctx context.Context, username string) (*biz.AuthUser, error) {
	var user model.User
	if err := r.data.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return toBizAuthUser(&user), nil
}

func (r *authRepo) FindByID(ctx context.Context, id uint64) (*biz.AuthUser, error) {
	var user model.User
	if err := r.data.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return toBizAuthUser(&user), nil
}

func (r *authRepo) CreateUser(ctx context.Context, user *biz.AuthUser) (*biz.AuthUser, error) {
	m := &model.User{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Password:    user.Password,
		Role:        user.Role,
	}
	if err := r.data.DB.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return toBizAuthUser(m), nil
}

func (r *authRepo) CreateChannel(ctx context.Context, userID uint64) error {
	ch := &model.Channel{
		UserID:     userID,
		MonthlyFee: 0,
	}
	return r.data.DB.WithContext(ctx).Create(ch).Error
}

func toBizAuthUser(m *model.User) *biz.AuthUser {
	return &biz.AuthUser{
		ID:          m.ID,
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Password:    m.Password,
		Role:        m.Role,
		IsHidden:    m.IsHidden,
		CreatedAt:   m.CreatedAt,
	}
}
