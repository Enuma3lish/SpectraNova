package data

import (
	"backend/internal/data/model"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

const (
	viewFlushInterval = 30 * time.Second
	cleanupInterval   = 10 * time.Second
)

// StartBackgroundWorkers launches cache maintenance goroutines.
// Called from NewData after all resources are initialized.
//
// Two workers run continuously:
//  1. View flush ticker — every 30s, drains views:buffer → batch UPDATE MySQL
//  2. Cleanup worker  — retries failed cache evictions from cleanup:queue
//
// Why Redis LIST for cleanup queue (not Go channel): survives app restarts.
// Why background goroutine for views (not synchronous): decouples write latency
// from user request latency.
func StartBackgroundWorkers(ctx context.Context, d *Data, logger log.Logger) {
	l := log.NewHelper(logger)

	if d.Redis == nil {
		l.Warn("background workers skipped: Redis not available")
		return
	}

	// Worker 1: flush view counts from Redis buffer to MySQL
	go func() {
		ticker := time.NewTicker(viewFlushInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				l.Info("view flush worker stopped")
				return
			case <-ticker.C:
				flushViewBuffer(ctx, d, l)
			}
		}
	}()

	// Worker 2: retry failed cache evictions
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				l.Info("cleanup worker stopped")
				return
			case <-ticker.C:
				processCleanupQueue(ctx, d, l)
			}
		}
	}()

	l.Info("background workers started (view flush, cleanup)")
}

// flushViewBuffer reads all buffered view increments from Redis and applies them
// to MySQL in a single transaction.
//
// Redis key: views:buffer (HASH)
// Fields: "{videoID}:member" or "{videoID}:non_member" → count
//
// After successful flush, deletes the buffer. If MySQL fails, the buffer is
// preserved and retried on the next tick.
func flushViewBuffer(ctx context.Context, d *Data, l *log.Helper) {
	counts, err := d.Redis.HGetAll(ctx, viewsBufferKey).Result()
	if err != nil || len(counts) == 0 {
		return
	}

	type viewUpdate struct {
		videoID  uint64
		member   int64
		nonMember int64
	}

	updates := make(map[uint64]*viewUpdate)

	for field, countStr := range counts {
		parts := strings.SplitN(field, ":", 2)
		if len(parts) != 2 {
			continue
		}
		videoID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			continue
		}
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			continue
		}

		vu, ok := updates[videoID]
		if !ok {
			vu = &viewUpdate{videoID: videoID}
			updates[videoID] = vu
		}
		if parts[1] == "member" {
			vu.member = count
		} else {
			vu.nonMember = count
		}
	}

	// Batch update MySQL in a single transaction
	err = d.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, vu := range updates {
			if vu.member > 0 {
				if err := tx.Model(&model.Video{}).Where("id = ?", vu.videoID).
					Update("views_member", gorm.Expr("views_member + ?", vu.member)).Error; err != nil {
					return err
				}
			}
			if vu.nonMember > 0 {
				if err := tx.Model(&model.Video{}).Where("id = ?", vu.videoID).
					Update("views_non_member", gorm.Expr("views_non_member + ?", vu.nonMember)).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		l.Warnf("view flush failed (will retry): %v", err)
		return
	}

	// Only delete buffer after successful MySQL flush
	d.Redis.Del(ctx, viewsBufferKey)
	l.Debugf("flushed %d view updates to MySQL", len(updates))
}

// processCleanupQueue retries failed cache evictions stored in the cleanup:queue LIST.
// Each job is formatted as "evict:{videoID}". On failure, the job is re-queued.
func processCleanupQueue(ctx context.Context, d *Data, l *log.Helper) {
	for {
		// LPOP one job at a time (non-blocking check)
		job, err := d.Redis.LPop(ctx, cleanupQueue).Result()
		if err != nil {
			return // queue empty or Redis error
		}

		parts := strings.SplitN(job, ":", 2)
		if len(parts) != 2 || parts[0] != "evict" {
			continue
		}

		videoID, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}

		// Look up tag IDs for this video from MySQL
		var tagIDs []uint64
		d.DB.Table("video_tags").Where("video_id = ?", videoID).Pluck("tag_id", &tagIDs)

		videoKey := fmt.Sprintf("%s%d", cacheVideoKeyPrefix, videoID)
		pipe := d.Redis.Pipeline()
		pipe.Del(ctx, videoKey)
		pipe.ZRem(ctx, popularKey, videoID)
		for _, tagID := range tagIDs {
			tagKey := fmt.Sprintf("%s%d", cacheTagKeyPrefix, tagID)
			pipe.SRem(ctx, tagKey, videoID)
		}

		if _, err := pipe.Exec(ctx); err != nil {
			l.Warnf("cleanup retry failed for video %d, re-queuing: %v", videoID, err)
			d.Redis.RPush(ctx, cleanupQueue, job) // re-queue at tail
			return // back off, retry on next tick
		}
	}
}
