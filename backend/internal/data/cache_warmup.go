package data

import (
	"backend/internal/data/model"
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	cacheTagKeyPrefix   = "tag:"
	cacheVideoKeyPrefix = "video:"
	cacheTagTTL         = 30 * time.Minute
	cacheVideoTTL       = 30 * time.Minute
)

// WarmUpCache loads all public, published, non-hidden videos into Redis
// on app boot. This eliminates cold start: the first user gets a cache HIT.
//
// Called from NewData() before servers start accepting traffic.
// Safe to call multiple times (idempotent — overwrites existing keys).
func (d *Data) WarmUpCache(ctx context.Context, logger log.Logger) {
	l := log.NewHelper(logger)

	if d.Redis == nil {
		l.Warn("cache warm-up skipped: Redis not available")
		return
	}

	// 1. Get all tags
	var tags []model.Tag
	if err := d.DB.Find(&tags).Error; err != nil {
		l.Warnf("cache warm-up: failed to query tags: %v", err)
		return
	}

	if len(tags) == 0 {
		l.Info("cache warm-up: no tags found, skipping")
		return
	}

	videosCached := 0

	for _, tag := range tags {
		tagKey := fmt.Sprintf("%s%d", cacheTagKeyPrefix, tag.ID)

		// 2. Query public, published, non-hidden videos for this tag
		var videoIDs []uint64
		err := d.DB.Table("video_tags").
			Select("video_tags.video_id").
			Joins("INNER JOIN videos ON videos.id = video_tags.video_id").
			Where("video_tags.tag_id = ?", tag.ID).
			Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false).
			Where("videos.access_tier = 0").
			Pluck("video_id", &videoIDs).Error
		if err != nil {
			l.Warnf("cache warm-up: failed to query videos for tag %d: %v", tag.ID, err)
			continue
		}

		if len(videoIDs) == 0 {
			continue
		}

		// 3. Populate tag SET
		members := make([]interface{}, len(videoIDs))
		for i, id := range videoIDs {
			members[i] = id
		}
		d.Redis.SAdd(ctx, tagKey, members...)
		d.Redis.Expire(ctx, tagKey, cacheTagTTL)

		// 4. Populate video HASHes (only if not already cached)
		for _, videoID := range videoIDs {
			videoKey := fmt.Sprintf("%s%d", cacheVideoKeyPrefix, videoID)

			// Skip if already cached (from another tag's iteration)
			exists, _ := d.Redis.Exists(ctx, videoKey).Result()
			if exists > 0 {
				continue
			}

			var video model.Video
			if err := d.DB.First(&video, videoID).Error; err != nil {
				continue
			}

			thumbnail := ""
			if video.ThumbnailURL != nil {
				thumbnail = *video.ThumbnailURL
			}

			d.Redis.HSet(ctx, videoKey, map[string]interface{}{
				"id":          video.ID,
				"title":       video.Title,
				"duration":    video.Duration,
				"views":       video.ViewsMember + video.ViewsNonMember,
				"thumbnail":   thumbnail,
				"category_id": video.CategoryID,
				"user_id":     video.UserID,
				"video_url":   video.VideoURL,
			})
			d.Redis.Expire(ctx, videoKey, cacheVideoTTL)
			videosCached++
		}
	}

	l.Infof("cache warm-up complete: %d tags, %d videos loaded into Redis", len(tags), videosCached)
}
