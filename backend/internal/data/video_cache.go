package data

import (
	"backend/internal/biz"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

const (
	popularKey     = "popular:global"
	viewsBufferKey = "views:buffer"
	cleanupQueue   = "cleanup:queue"
	popularTTL     = 10 * time.Minute
)

// VideoCache provides Redis cache operations for the recommendation system.
// Uses a two-layer design:
//   - Index layer: SET per tag containing video IDs (tag:{id})
//   - Data layer:  HASH per video containing summary fields (video:{id})
//
// Why two layers: a video with 3 tags appears in 3 tag SETs, but its data
// is stored only once in a HASH. This avoids duplication and makes updates
// easier (one HASH to update vs N copies).
type VideoCache struct {
	data *Data
	log  *log.Helper
}

func NewVideoCache(data *Data, logger log.Logger) *VideoCache {
	return &VideoCache{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetRecommended attempts to serve recommendations from Redis cache.
// Returns nil if cache is unavailable or empty (caller should fall back to MySQL).
//
// Algorithm:
// 1. SUNIONSTORE temp → merge video IDs from all selected tag SETs (OR, not AND)
// 2. Shuffle the merged IDs → take page_size for variety
// 3. Pipeline HGETALL for each video ID → return hydrated results
func (vc *VideoCache) GetRecommended(ctx context.Context, tagIDs []uint64, offset, limit int) ([]*biz.Video, int64, error) {
	if vc.data.Redis == nil || len(tagIDs) == 0 {
		return nil, 0, nil
	}

	// Build tag key names
	tagKeys := make([]string, len(tagIDs))
	for i, id := range tagIDs {
		tagKeys[i] = fmt.Sprintf("%s%d", cacheTagKeyPrefix, id)
	}

	// SUNION: merge all video IDs across selected tags (OR semantics = more variety)
	videoIDStrs, err := vc.data.Redis.SUnion(ctx, tagKeys...).Result()
	if err != nil || len(videoIDStrs) == 0 {
		return nil, 0, nil // cache miss → caller falls back to DB
	}

	total := int64(len(videoIDStrs))

	// Shuffle for randomness, then paginate
	rand.Shuffle(len(videoIDStrs), func(i, j int) {
		videoIDStrs[i], videoIDStrs[j] = videoIDStrs[j], videoIDStrs[i]
	})

	// Apply pagination
	if offset >= len(videoIDStrs) {
		return []*biz.Video{}, total, nil
	}
	end := offset + limit
	if end > len(videoIDStrs) {
		end = len(videoIDStrs)
	}
	page := videoIDStrs[offset:end]

	// Pipeline HGETALL for each video (one round trip for all)
	pipe := vc.data.Redis.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(page))
	for i, idStr := range page {
		videoKey := fmt.Sprintf("%s%s", cacheVideoKeyPrefix, idStr)
		cmds[i] = pipe.HGetAll(ctx, videoKey)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, 0, nil
	}

	videos := make([]*biz.Video, 0, len(page))
	for _, cmd := range cmds {
		m, err := cmd.Result()
		if err != nil || len(m) == 0 {
			continue // cache miss for this video, skip
		}
		v := hashToVideo(m)
		if v != nil {
			videos = append(videos, v)
		}
	}

	if len(videos) == 0 {
		return nil, 0, nil // all cache misses
	}

	return videos, total, nil
}

// CacheVideo writes a video's summary into Redis (both tag SETs and video HASH).
// Called after video creation or lazy-populate on cache miss.
func (vc *VideoCache) CacheVideo(ctx context.Context, v *biz.Video, tagIDs []uint64) {
	if vc.data.Redis == nil || v == nil {
		return
	}

	videoKey := fmt.Sprintf("%s%d", cacheVideoKeyPrefix, v.ID)
	pipe := vc.data.Redis.Pipeline()

	// Video HASH
	pipe.HSet(ctx, videoKey, map[string]interface{}{
		"id":          v.ID,
		"title":       v.Title,
		"duration":    v.Duration,
		"views":       v.ViewsMember + v.ViewsNonMember,
		"thumbnail":   v.ThumbnailURL,
		"category_id": v.CategoryID,
		"user_id":     v.UserID,
		"video_url":   v.VideoURL,
		"username":    v.Username,
		"created_at":  v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
	pipe.Expire(ctx, videoKey, cacheVideoTTL)

	// Add to each tag SET
	for _, tagID := range tagIDs {
		tagKey := fmt.Sprintf("%s%d", cacheTagKeyPrefix, tagID)
		pipe.SAdd(ctx, tagKey, v.ID)
		pipe.Expire(ctx, tagKey, cacheTagTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		vc.log.Warnf("failed to cache video %d: %v", v.ID, err)
	}
}

// EvictVideo removes a video from all tag SETs and deletes its HASH.
// Called on video update, delete, or admin hide.
// On failure, pushes to cleanup queue for retry.
func (vc *VideoCache) EvictVideo(ctx context.Context, videoID uint64, tagIDs []uint64) {
	if vc.data.Redis == nil {
		return
	}

	videoKey := fmt.Sprintf("%s%d", cacheVideoKeyPrefix, videoID)
	pipe := vc.data.Redis.Pipeline()

	pipe.Del(ctx, videoKey)
	pipe.ZRem(ctx, popularKey, videoID)

	for _, tagID := range tagIDs {
		tagKey := fmt.Sprintf("%s%d", cacheTagKeyPrefix, tagID)
		pipe.SRem(ctx, tagKey, videoID)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		vc.log.Warnf("failed to evict video %d from cache, queuing for retry: %v", videoID, err)
		// Push to cleanup queue for background retry
		job := fmt.Sprintf("evict:%d", videoID)
		vc.data.Redis.RPush(ctx, cleanupQueue, job)
	}
}

// IncrementViewsBuffered buffers a view increment in Redis instead of hitting MySQL directly.
// Why buffer: 1000 concurrent views → 1000 MySQL UPDATEs → DB overload.
// Buffering reduces to 1 UPDATE per 30s per video.
func (vc *VideoCache) IncrementViewsBuffered(ctx context.Context, videoID uint64, isMember bool) {
	if vc.data.Redis == nil {
		return
	}

	suffix := "non_member"
	if isMember {
		suffix = "member"
	}
	field := fmt.Sprintf("%d:%s", videoID, suffix)
	vc.data.Redis.HIncrBy(ctx, viewsBufferKey, field, 1)

	// Also update popular ranking instantly
	vc.data.Redis.ZIncrBy(ctx, popularKey, 1, strconv.FormatUint(videoID, 10))
}

// hashToVideo converts a Redis HASH map to a biz.Video.
func hashToVideo(m map[string]string) *biz.Video {
	id, err := strconv.ParseUint(m["id"], 10, 64)
	if err != nil {
		return nil
	}
	duration, _ := strconv.ParseUint(m["duration"], 10, 32)
	views, _ := strconv.ParseUint(m["views"], 10, 64)
	categoryID, _ := strconv.ParseUint(m["category_id"], 10, 64)
	userID, _ := strconv.ParseUint(m["user_id"], 10, 64)
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", m["created_at"])

	return &biz.Video{
		ID:             id,
		UserID:         userID,
		Username:       m["username"],
		CategoryID:     categoryID,
		Title:          m["title"],
		ThumbnailURL:   m["thumbnail"],
		VideoURL:       m["video_url"],
		Duration:       uint32(duration),
		ViewsNonMember: views, // combined in cache, stored in one field
		IsPublished:    true,
		CreatedAt:      createdAt,
	}
}
