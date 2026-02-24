package biz

import (
	"context"
	"time"

	"backend/internal/pkg/pagination"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type Video struct {
	ID             uint64
	UserID         uint64
	Username       string
	CategoryID     uint64
	CategoryName   string
	Title          string
	Description    string
	VideoURL       string
	ThumbnailURL   string
	Duration       uint32
	ViewsMember    uint64
	ViewsNonMember uint64
	AccessTier     int8
	IsPublished    bool
	IsHidden       bool
	Tags           []*Tag
	CreatedAt      time.Time
}

type VideoRepo interface {
	Create(ctx context.Context, video *Video) (*Video, error)
	Update(ctx context.Context, video *Video) (*Video, error)
	Delete(ctx context.Context, id uint64) error
	FindByID(ctx context.Context, id uint64) (*Video, error)
	ListByTags(ctx context.Context, tagIDs []uint64, offset, limit int) ([]*Video, int64, error)
	ListRandom(ctx context.Context, offset, limit int) ([]*Video, int64, error)
	IncrementViews(ctx context.Context, id uint64, isMember bool) error
	TogglePublish(ctx context.Context, id uint64, published bool) error
	GetTagIDsByVideo(ctx context.Context, videoID uint64) ([]uint64, error)
	SetVideoTags(ctx context.Context, videoID uint64, tagIDs []uint64) error
}

// MembershipChecker checks if a user has a membership to a channel.
// Implemented by ChannelRepo in the data layer.
type MembershipChecker interface {
	HasMembership(ctx context.Context, userID, channelOwnerUserID uint64) (tier int8, err error)
}

type VideoUsecase struct {
	repo       VideoRepo
	tagUsecase *TagUsecase
	membership MembershipChecker
	log        *log.Helper
}

func NewVideoUsecase(repo VideoRepo, tagUsecase *TagUsecase, membership MembershipChecker, logger log.Logger) *VideoUsecase {
	return &VideoUsecase{
		repo:       repo,
		tagUsecase: tagUsecase,
		membership: membership,
		log:        log.NewHelper(logger),
	}
}

func (uc *VideoUsecase) CreateVideo(ctx context.Context, userID uint64, video *Video) (*Video, error) {
	video.UserID = userID
	video.IsPublished = true
	video.IsHidden = false

	created, err := uc.repo.Create(ctx, video)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL", "failed to create video")
	}

	// Associate tags
	if len(video.Tags) > 0 {
		tagIDs := make([]uint64, len(video.Tags))
		for i, t := range video.Tags {
			tagIDs[i] = t.ID
		}
		if err := uc.repo.SetVideoTags(ctx, created.ID, tagIDs); err != nil {
			uc.log.Warnf("failed to set video tags: %v", err)
		}
	}

	return uc.repo.FindByID(ctx, created.ID)
}

func (uc *VideoUsecase) GetVideo(ctx context.Context, videoID uint64, viewerID *uint64, viewerRole string) (*Video, error) {
	video, err := uc.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}

	isOwner := viewerID != nil && *viewerID == video.UserID
	isAdmin := viewerRole == "admin"

	// Hidden check: only admin or owner can see
	if video.IsHidden && !isAdmin && !isOwner {
		return nil, errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}

	// Published check: only owner can see unpublished
	if !video.IsPublished && !isOwner {
		return nil, errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}

	// Access tier check
	if video.AccessTier > 0 && !isOwner && !isAdmin {
		if viewerID == nil {
			return nil, errors.Forbidden("VIDEO_ACCESS_DENIED", "membership required")
		}
		if uc.membership != nil {
			tier, err := uc.membership.HasMembership(ctx, *viewerID, video.UserID)
			if err != nil || tier < video.AccessTier {
				return nil, errors.Forbidden("VIDEO_ACCESS_DENIED", "insufficient membership tier")
			}
		}
	}

	// Increment views
	isMember := viewerID != nil
	_ = uc.repo.IncrementViews(ctx, videoID, isMember)

	return video, nil
}

func (uc *VideoUsecase) UpdateVideo(ctx context.Context, userID uint64, video *Video) (*Video, error) {
	existing, err := uc.repo.FindByID(ctx, video.ID)
	if err != nil {
		return nil, errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}
	if existing.UserID != userID {
		return nil, errors.Forbidden("VIDEO_NOT_OWNER", "not the owner of this video")
	}

	updated, err := uc.repo.Update(ctx, video)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL", "failed to update video")
	}

	// Update tags if provided
	if len(video.Tags) > 0 {
		tagIDs := make([]uint64, len(video.Tags))
		for i, t := range video.Tags {
			tagIDs[i] = t.ID
		}
		_ = uc.repo.SetVideoTags(ctx, video.ID, tagIDs)
	}

	return uc.repo.FindByID(ctx, updated.ID)
}

func (uc *VideoUsecase) DeleteVideo(ctx context.Context, userID, videoID uint64) error {
	video, err := uc.repo.FindByID(ctx, videoID)
	if err != nil {
		return errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}
	if video.UserID != userID {
		return errors.Forbidden("VIDEO_NOT_OWNER", "not the owner of this video")
	}
	if video.IsPublished {
		return errors.BadRequest("VIDEO_ACCESS_DENIED", "cannot delete a published video; unpublish first")
	}
	return uc.repo.Delete(ctx, videoID)
}

func (uc *VideoUsecase) TogglePublish(ctx context.Context, userID, videoID uint64, published bool) (*Video, error) {
	video, err := uc.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, errors.NotFound("VIDEO_NOT_FOUND", "video not found")
	}
	if video.UserID != userID {
		return nil, errors.Forbidden("VIDEO_NOT_OWNER", "not the owner of this video")
	}

	if err := uc.repo.TogglePublish(ctx, videoID, published); err != nil {
		return nil, errors.InternalServer("INTERNAL", "failed to toggle publish")
	}

	return uc.repo.FindByID(ctx, videoID)
}

func (uc *VideoUsecase) GetRecommended(ctx context.Context, userID *uint64, sessionID *string, page, pageSize int32) ([]*Video, int64, error) {
	offset, limit := pagination.Normalize(page, pageSize)

	tagIDs, err := uc.tagUsecase.GetRecommendedTagIDs(ctx, userID, sessionID)
	if err != nil || len(tagIDs) == 0 {
		// No tags: fallback to random published videos
		return uc.repo.ListRandom(ctx, offset, limit)
	}

	return uc.repo.ListByTags(ctx, tagIDs, offset, limit)
}
