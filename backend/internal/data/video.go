package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type videoRepo struct {
	data  *Data
	cache *VideoCache
	log   *log.Helper
}

func NewVideoRepo(data *Data, cache *VideoCache, logger log.Logger) biz.VideoRepo {
	return &videoRepo{
		data:  data,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

func (r *videoRepo) Create(ctx context.Context, video *biz.Video) (*biz.Video, error) {
	var desc *string
	if video.Description != "" {
		desc = &video.Description
	}
	var thumb *string
	if video.ThumbnailURL != "" {
		thumb = &video.ThumbnailURL
	}
	m := &model.Video{
		UserID:       video.UserID,
		CategoryID:   video.CategoryID,
		Title:        video.Title,
		Description:  desc,
		VideoURL:     video.VideoURL,
		ThumbnailURL: thumb,
		Duration:     video.Duration,
		AccessTier:   video.AccessTier,
		IsPublished:  video.IsPublished,
		IsHidden:     false,
	}
	if err := r.data.DB.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return r.FindByID(ctx, m.ID)
}

func (r *videoRepo) Update(ctx context.Context, video *biz.Video) (*biz.Video, error) {
	updates := map[string]interface{}{}
	if video.Title != "" {
		updates["title"] = video.Title
	}
	if video.Description != "" {
		updates["description"] = video.Description
	}
	if video.CategoryID != 0 {
		updates["category_id"] = video.CategoryID
	}
	if video.ThumbnailURL != "" {
		updates["thumbnail_url"] = video.ThumbnailURL
	}
	if video.AccessTier >= 0 {
		updates["access_tier"] = video.AccessTier
	}

	if len(updates) > 0 {
		if err := r.data.DB.WithContext(ctx).Model(&model.Video{}).Where("id = ?", video.ID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return r.FindByID(ctx, video.ID)
}

func (r *videoRepo) Delete(ctx context.Context, id uint64) error {
	return r.data.DB.WithContext(ctx).Delete(&model.Video{}, id).Error
}

func (r *videoRepo) FindByID(ctx context.Context, id uint64) (*biz.Video, error) {
	var video model.Video
	if err := r.data.DB.WithContext(ctx).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		First(&video, id).Error; err != nil {
		return nil, err
	}
	return toBizVideo(&video), nil
}

func (r *videoRepo) ListByTags(ctx context.Context, tagIDs []uint64, offset, limit int) ([]*biz.Video, int64, error) {
	// Try cache first (sub-millisecond vs ~50ms MySQL)
	if r.cache != nil {
		videos, total, err := r.cache.GetRecommended(ctx, tagIDs, offset, limit)
		if err == nil && len(videos) > 0 {
			return videos, total, nil
		}
	}

	// Cache miss: fall back to MySQL
	var total int64
	var videos []model.Video

	baseQuery := r.data.DB.WithContext(ctx).
		Model(&model.Video{}).
		Joins("INNER JOIN video_tags ON video_tags.video_id = videos.id").
		Where("video_tags.tag_id IN ?", tagIDs).
		Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false).
		Where("videos.access_tier = 0").
		Group("videos.id")

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.data.DB.WithContext(ctx).
		Preload("Tags").Preload("Category").Preload("User").
		Joins("INNER JOIN video_tags ON video_tags.video_id = videos.id").
		Where("video_tags.tag_id IN ?", tagIDs).
		Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false).
		Where("videos.access_tier = 0").
		Group("videos.id").
		Order("RAND()").
		Offset(offset).Limit(limit).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return toBizVideos(videos), total, nil
}

func (r *videoRepo) ListRandom(ctx context.Context, offset, limit int) ([]*biz.Video, int64, error) {
	var total int64
	var videos []model.Video

	baseQuery := r.data.DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("is_published = ? AND is_hidden = ? AND deleted_at IS NULL", true, false).
		Where("access_tier = 0")

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.data.DB.WithContext(ctx).
		Preload("Tags").Preload("Category").Preload("User").
		Where("is_published = ? AND is_hidden = ? AND deleted_at IS NULL", true, false).
		Where("access_tier = 0").
		Order("RAND()").
		Offset(offset).Limit(limit).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return toBizVideos(videos), total, nil
}

func (r *videoRepo) IncrementViews(ctx context.Context, id uint64, isMember bool) error {
	// Buffer through Redis (flushed to MySQL every 30s by background worker).
	// Direct MySQL UPDATE would create hot-row contention under load.
	if r.cache != nil {
		r.cache.IncrementViewsBuffered(ctx, id, isMember)
		return nil
	}
	// Fallback: direct MySQL if Redis unavailable
	col := "views_non_member"
	if isMember {
		col = "views_member"
	}
	return r.data.DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("id = ?", id).
		Update(col, gorm.Expr(col+" + 1")).Error
}

func (r *videoRepo) TogglePublish(ctx context.Context, id uint64, published bool) error {
	return r.data.DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("id = ?", id).
		Update("is_published", published).Error
}

func (r *videoRepo) GetTagIDsByVideo(ctx context.Context, videoID uint64) ([]uint64, error) {
	var video model.Video
	if err := r.data.DB.WithContext(ctx).Preload("Tags").First(&video, videoID).Error; err != nil {
		return nil, err
	}
	ids := make([]uint64, len(video.Tags))
	for i, t := range video.Tags {
		ids[i] = t.ID
	}
	return ids, nil
}

func (r *videoRepo) SetVideoTags(ctx context.Context, videoID uint64, tagIDs []uint64) error {
	var video model.Video
	if err := r.data.DB.WithContext(ctx).First(&video, videoID).Error; err != nil {
		return err
	}
	tags := make([]model.Tag, len(tagIDs))
	for i, id := range tagIDs {
		tags[i] = model.Tag{ID: id}
	}
	return r.data.DB.WithContext(ctx).Model(&video).Association("Tags").Replace(tags)
}

func toBizVideo(m *model.Video) *biz.Video {
	desc := ""
	if m.Description != nil {
		desc = *m.Description
	}
	thumb := ""
	if m.ThumbnailURL != nil {
		thumb = *m.ThumbnailURL
	}
	username := ""
	categoryName := ""
	if m.User.ID != 0 {
		username = m.User.DisplayName
	}
	if m.Category.ID != 0 {
		categoryName = m.Category.Name
	}

	tags := make([]*biz.Tag, len(m.Tags))
	for i, t := range m.Tags {
		tags[i] = &biz.Tag{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}

	return &biz.Video{
		ID:             m.ID,
		UserID:         m.UserID,
		Username:       username,
		CategoryID:     m.CategoryID,
		CategoryName:   categoryName,
		Title:          m.Title,
		Description:    desc,
		VideoURL:       m.VideoURL,
		ThumbnailURL:   thumb,
		Duration:       m.Duration,
		ViewsMember:    m.ViewsMember,
		ViewsNonMember: m.ViewsNonMember,
		AccessTier:     m.AccessTier,
		IsPublished:    m.IsPublished,
		IsHidden:       m.IsHidden,
		Tags:           tags,
		CreatedAt:      m.CreatedAt,
	}
}

func toBizVideos(models []model.Video) []*biz.Video {
	videos := make([]*biz.Video, len(models))
	for i := range models {
		videos[i] = toBizVideo(&models[i])
	}
	return videos
}
