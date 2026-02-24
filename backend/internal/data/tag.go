package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type tagRepo struct {
	data *Data
	log  *log.Helper
}

func NewTagRepo(data *Data, logger log.Logger) biz.TagRepo {
	return &tagRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *tagRepo) ListAll(ctx context.Context) ([]*biz.Tag, error) {
	var tags []model.Tag
	if err := r.data.DB.WithContext(ctx).Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return toBizTags(tags), nil
}

func (r *tagRepo) GetUserTags(ctx context.Context, userID *uint64, sessionID *string) ([]*biz.Tag, error) {
	var prefs []model.UserTagPreference
	q := r.data.DB.WithContext(ctx).Preload("Tag")
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	} else if sessionID != nil {
		q = q.Where("session_id = ?", *sessionID)
	} else {
		return nil, nil
	}
	if err := q.Find(&prefs).Error; err != nil {
		return nil, err
	}

	tags := make([]*biz.Tag, len(prefs))
	for i, p := range prefs {
		tags[i] = &biz.Tag{
			ID:   p.Tag.ID,
			Name: p.Tag.Name,
			Slug: p.Tag.Slug,
		}
	}
	return tags, nil
}

func (r *tagRepo) SetUserTags(ctx context.Context, userID *uint64, sessionID *string, tagIDs []uint64) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing preferences
		q := tx.Where("1 = 1")
		if userID != nil {
			q = q.Where("user_id = ?", *userID)
		} else {
			q = q.Where("session_id = ?", *sessionID)
		}
		if err := q.Delete(&model.UserTagPreference{}).Error; err != nil {
			return err
		}

		// Insert new preferences
		for _, tagID := range tagIDs {
			pref := model.UserTagPreference{
				UserID:    userID,
				TagID:     tagID,
				SessionID: sessionID,
			}
			if err := tx.Create(&pref).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *tagRepo) GetTagsByIDs(ctx context.Context, ids []uint64) ([]*biz.Tag, error) {
	var tags []model.Tag
	if err := r.data.DB.WithContext(ctx).Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, err
	}
	return toBizTags(tags), nil
}

func toBizTags(models []model.Tag) []*biz.Tag {
	tags := make([]*biz.Tag, len(models))
	for i, m := range models {
		tags[i] = &biz.Tag{
			ID:   m.ID,
			Name: m.Name,
			Slug: m.Slug,
		}
	}
	return tags
}
