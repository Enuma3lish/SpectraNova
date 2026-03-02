package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type adminRepo struct {
	data *Data
	log  *log.Helper
}

func NewAdminRepo(data *Data, logger log.Logger) biz.AdminRepo {
	return &adminRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *adminRepo) ListUsers(ctx context.Context, offset, limit int) ([]*biz.AdminUser, int64, error) {
	var users []model.User
	var total int64

	db := r.data.DB.WithContext(ctx).Model(&model.User{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.AdminUser, len(users))
	for i, u := range users {
		result[i] = toBizAdminUser(&u)
	}
	return result, total, nil
}

func (r *adminRepo) FindUserByID(ctx context.Context, id uint64) (*biz.AdminUser, error) {
	var user model.User
	if err := r.data.DB.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return toBizAdminUser(&user), nil
}

func (r *adminRepo) DeleteUser(ctx context.Context, id uint64) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete related records first
		if err := tx.Where("user_id = ?", id).Delete(&model.Membership{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.UserTagPreference{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.ViewRecord{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.Notification{}).Error; err != nil {
			return err
		}
		// Delete donations (both sent and received)
		if err := tx.Where("from_user_id = ? OR to_user_id = ?", id, id).Delete(&model.Donation{}).Error; err != nil {
			return err
		}
		// Delete user's videos
		if err := tx.Where("user_id = ?", id).Delete(&model.Video{}).Error; err != nil {
			return err
		}
		// Delete channel
		if err := tx.Where("user_id = ?", id).Delete(&model.Channel{}).Error; err != nil {
			return err
		}
		// Delete user
		if err := tx.Delete(&model.User{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *adminRepo) ListAllVideos(ctx context.Context, offset, limit int) ([]*biz.AdminVideo, int64, error) {
	var videos []model.Video
	var total int64

	db := r.data.DB.WithContext(ctx).Model(&model.Video{}).Unscoped()
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("User").Preload("Category").
		Offset(offset).Limit(limit).Order("id DESC").
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*biz.AdminVideo, len(videos))
	for i, v := range videos {
		result[i] = toBizAdminVideo(&v)
	}
	return result, total, nil
}

func (r *adminRepo) DeleteVideo(ctx context.Context, id uint64) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove video_tags associations
		if err := tx.Exec("DELETE FROM video_tags WHERE video_id = ?", id).Error; err != nil {
			return err
		}
		// Delete view records
		if err := tx.Where("video_id = ?", id).Delete(&model.ViewRecord{}).Error; err != nil {
			return err
		}
		// Delete donations for this video
		if err := tx.Where("video_id = ?", id).Delete(&model.Donation{}).Error; err != nil {
			return err
		}
		// Hard delete the video (Unscoped bypasses soft delete)
		if err := tx.Unscoped().Delete(&model.Video{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *adminRepo) CreateTag(ctx context.Context, tag *biz.AdminTag) (*biz.AdminTag, error) {
	m := &model.Tag{
		Name: tag.Name,
		Slug: tag.Slug,
	}
	if err := r.data.DB.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return &biz.AdminTag{ID: m.ID, Name: m.Name, Slug: m.Slug}, nil
}

func (r *adminRepo) UpdateTag(ctx context.Context, tag *biz.AdminTag) (*biz.AdminTag, error) {
	if err := r.data.DB.WithContext(ctx).Model(&model.Tag{}).Where("id = ?", tag.ID).
		Updates(map[string]interface{}{
			"name": tag.Name,
			"slug": tag.Slug,
		}).Error; err != nil {
		return nil, err
	}
	return r.FindTagByID(ctx, tag.ID)
}

func (r *adminRepo) DeleteTag(ctx context.Context, id uint64) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove video_tags associations
		if err := tx.Exec("DELETE FROM video_tags WHERE tag_id = ?", id).Error; err != nil {
			return err
		}
		// Remove user tag preferences
		if err := tx.Where("tag_id = ?", id).Delete(&model.UserTagPreference{}).Error; err != nil {
			return err
		}
		// Delete tag
		if err := tx.Delete(&model.Tag{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *adminRepo) FindTagByID(ctx context.Context, id uint64) (*biz.AdminTag, error) {
	var tag model.Tag
	if err := r.data.DB.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &biz.AdminTag{ID: tag.ID, Name: tag.Name, Slug: tag.Slug}, nil
}

func (r *adminRepo) FindTagByName(ctx context.Context, name string) (*biz.AdminTag, error) {
	var tag model.Tag
	if err := r.data.DB.WithContext(ctx).Where("name = ?", name).First(&tag).Error; err != nil {
		return nil, err
	}
	return &biz.AdminTag{ID: tag.ID, Name: tag.Name, Slug: tag.Slug}, nil
}

func toBizAdminUser(m *model.User) *biz.AdminUser {
	return &biz.AdminUser{
		ID:          m.ID,
		Username:    m.Username,
		DisplayName: m.DisplayName,
		Role:        m.Role,
		IsHidden:    m.IsHidden,
		CreatedAt:   m.CreatedAt,
	}
}

func toBizAdminVideo(m *model.Video) *biz.AdminVideo {
	v := &biz.AdminVideo{
		ID:             m.ID,
		Title:          m.Title,
		UserID:         m.UserID,
		AccessTier:     m.AccessTier,
		IsPublished:    m.IsPublished,
		IsHidden:       m.IsHidden,
		ViewsMember:    m.ViewsMember,
		ViewsNonMember: m.ViewsNonMember,
		CreatedAt:      m.CreatedAt,
	}
	if m.User.Username != "" {
		v.Username = m.User.Username
	}
	if m.Category.Name != "" {
		v.CategoryName = m.Category.Name
	}
	return v
}
