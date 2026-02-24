package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type Channel struct {
	ID              uint64
	UserID          uint64
	DisplayName     string
	AvatarURL       string
	MonthlyFee      float64
	SubscriberCount int64
	IsHidden        bool
}

type Membership struct {
	ID        uint64
	ChannelID uint64
	UserID    uint64
	Tier      int8
	Status    string
}

type ChannelRepo interface {
	FindByID(ctx context.Context, id uint64) (*Channel, error)
	FindByUserID(ctx context.Context, userID uint64) (*Channel, error)
	GetSubscriberCount(ctx context.Context, channelID uint64) (int64, error)
	GetMembership(ctx context.Context, userID, channelID uint64) (*Membership, error)
	Subscribe(ctx context.Context, userID, channelID uint64) error
	Unsubscribe(ctx context.Context, userID, channelID uint64) error
	// HasMembership implements MembershipChecker for VideoService
	HasMembership(ctx context.Context, userID, channelOwnerUserID uint64) (tier int8, err error)
}

type ChannelUsecase struct {
	repo ChannelRepo
	log  *log.Helper
}

func NewChannelUsecase(repo ChannelRepo, logger log.Logger) *ChannelUsecase {
	return &ChannelUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *ChannelUsecase) GetChannel(ctx context.Context, channelID uint64, viewerID *uint64) (*Channel, string, error) {
	ch, err := uc.repo.FindByID(ctx, channelID)
	if err != nil {
		return nil, "", errors.NotFound("CHANNEL_NOT_FOUND", "channel not found")
	}

	count, _ := uc.repo.GetSubscriberCount(ctx, channelID)
	ch.SubscriberCount = count

	membershipStatus := "none"
	if viewerID != nil {
		m, _ := uc.repo.GetMembership(ctx, *viewerID, channelID)
		if m != nil && m.Status == "active" {
			if m.Tier == 2 {
				membershipStatus = "premium"
			} else {
				membershipStatus = "subscribed"
			}
		}
	}

	return ch, membershipStatus, nil
}

func (uc *ChannelUsecase) Subscribe(ctx context.Context, userID, channelID uint64) error {
	ch, err := uc.repo.FindByID(ctx, channelID)
	if err != nil {
		return errors.NotFound("CHANNEL_NOT_FOUND", "channel not found")
	}

	if ch.UserID == userID {
		return errors.BadRequest("CHANNEL_SELF_SUBSCRIBE", "cannot subscribe to your own channel")
	}

	existing, _ := uc.repo.GetMembership(ctx, userID, channelID)
	if existing != nil && existing.Status == "active" {
		return errors.Conflict("CHANNEL_ALREADY_SUBSCRIBED", "already subscribed")
	}

	return uc.repo.Subscribe(ctx, userID, channelID)
}

func (uc *ChannelUsecase) Unsubscribe(ctx context.Context, userID, channelID uint64) error {
	existing, _ := uc.repo.GetMembership(ctx, userID, channelID)
	if existing == nil {
		return errors.NotFound("CHANNEL_NOT_SUBSCRIBED", "not subscribed to this channel")
	}

	return uc.repo.Unsubscribe(ctx, userID, channelID)
}
