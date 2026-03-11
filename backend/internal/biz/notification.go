package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	NotificationTypeVideoLiked        = "video_liked"
	NotificationTypeModerationRemoved = "moderation_removed"
)

type NotificationEvent struct {
	Type        string `json:"type"`
	RecipientID uint64 `json:"recipient_id"`
	ActorID     uint64 `json:"actor_id,omitempty"`
	VideoID     uint64 `json:"video_id,omitempty"`
	Title       string `json:"title"`
	Message     string `json:"message"`
	Reason      string `json:"reason,omitempty"`
}

type NotificationPublisher interface {
	Publish(ctx context.Context, event *NotificationEvent) error
}

type NotificationUsecase struct {
	videos    VideoRepo
	publisher NotificationPublisher
	log       *log.Helper
}

func NewNotificationUsecase(videos VideoRepo, publisher NotificationPublisher, logger log.Logger) *NotificationUsecase {
	return &NotificationUsecase{
		videos:    videos,
		publisher: publisher,
		log:       log.NewHelper(logger),
	}
}

func (uc *NotificationUsecase) NotifyVideoLiked(ctx context.Context, viewerID, videoID uint64) error {
	video, err := uc.videos.FindByID(ctx, videoID)
	if err != nil {
		return errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}
	if video.UserID == viewerID {
		return nil
	}

	event := &NotificationEvent{
		Type:        NotificationTypeVideoLiked,
		RecipientID: video.UserID,
		ActorID:     viewerID,
		VideoID:     video.ID,
		Title:       "Your video got a new like",
		Message:     fmt.Sprintf("Someone liked your video \"%s\".", video.Title),
	}

	if err := uc.publisher.Publish(ctx, event); err != nil {
		uc.log.Warnf("failed to publish like notification for video %d: %v", videoID, err)
		return errors.InternalServer("INTERNAL", "failed to publish notification")
	}

	return nil
}

func (uc *NotificationUsecase) NotifyModerationRemoved(ctx context.Context, adminID uint64, video *Video, reason string) error {
	if video == nil {
		return errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}
	if reason == "" {
		reason = "Your work is illegal and has been removed by an administrator."
	}

	event := &NotificationEvent{
		Type:        NotificationTypeModerationRemoved,
		RecipientID: video.UserID,
		ActorID:     adminID,
		VideoID:     video.ID,
		Title:       "Your video was removed",
		Message:     reason,
		Reason:      reason,
	}

	if err := uc.publisher.Publish(ctx, event); err != nil {
		uc.log.Warnf("failed to publish moderation notification for video %d: %v", video.ID, err)
		return err
	}

	return nil
}
