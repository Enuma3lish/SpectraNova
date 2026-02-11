package data
package data

import (
	"context"
	"errors"

	"MLW/fenzVideo/internal/biz"
	"MLW/fenzVideo/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type AuthRepo struct {
	data *Data
	log  *log.Helper
}

func NewAuthRepo(data *Data, logger log.Logger) *AuthRepo {
	return &AuthRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user *biz.User) (*biz.User, error) {
	modelUser := model.User{
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		Role:         user.Role,
		IsHidden:     user.IsHidden,
	}
	if err := r.data.db.WithContext(ctx).Create(&modelUser).Error; err != nil {
		return nil, err
	}
	return toBizUser(&modelUser), nil
}

func (r *AuthRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var modelUser model.User
	err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&modelUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBizUser(&modelUser), nil
}

func (r *AuthRepo) GetByID(ctx context.Context, id int64) (*biz.User, error) {
	var modelUser model.User
	err := r.data.db.WithContext(ctx).First(&modelUser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}
	return toBizUser(&modelUser), nil
}

func toBizUser(user *model.User) *biz.User {
	return &biz.User{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		Role:         user.Role,
		IsHidden:     user.IsHidden,
	}
}














}	}		IsHidden:     user.IsHidden,		Role:         user.Role,		DisplayName:  user.DisplayName,		PasswordHash: user.PasswordHash,		Username:     user.Username,		ID:           user.ID,	return &biz.User{func toBizUser(user *model.User) *biz.User {}	return toBizUser(&modelUser), nil	}		return nil, err		}			return nil, biz.ErrUserNotFound		if errors.Is(err, gorm.ErrRecordNotFound) {	if err != nil {	err := r.data.db.WithContext(ctx).First(&modelUser, id).Error	var modelUser model.Userfunc (r *AuthRepo) GetByID(ctx context.Context, id int64) (*biz.User, error) {}	return toBizUser(&modelUser), nil	}		return nil, err		}			return nil, biz.ErrUserNotFound		if errors.Is(err, gorm.ErrRecordNotFound) {	if err != nil {	err := r.data.db.WithContext(ctx).Where("username = ?", username).First(&modelUser).Error	var modelUser model.Userfunc (r *AuthRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {}