package data

import (
	"context"
	"time"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type channelRepo struct {
	data *Data
	log  *log.Helper
}

func NewChannelRepo(data *Data, logger log.Logger) biz.ChannelRepo {
	return &channelRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *channelRepo) FindByID(ctx context.Context, id uint64) (*biz.Channel, error) {
	var ch model.Channel
	if err := r.data.DB.WithContext(ctx).Preload("User").First(&ch, id).Error; err != nil {
		return nil, err
	}
	return toBizChannel(&ch), nil
}

func (r *channelRepo) FindByUserID(ctx context.Context, userID uint64) (*biz.Channel, error) {
	var ch model.Channel
	if err := r.data.DB.WithContext(ctx).Preload("User").Where("user_id = ?", userID).First(&ch).Error; err != nil {
		return nil, err
	}
	return toBizChannel(&ch), nil
}

func (r *channelRepo) GetSubscriberCount(ctx context.Context, channelID uint64) (int64, error) {
	var count int64
	err := r.data.DB.WithContext(ctx).
		Model(&model.Membership{}).
		Where("channel_id = ? AND status = ?", channelID, "active").
		Count(&count).Error
	return count, err
}

func (r *channelRepo) GetMembership(ctx context.Context, userID, channelID uint64) (*biz.Membership, error) {
	var m model.Membership
	if err := r.data.DB.WithContext(ctx).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.Membership{
		ID:        m.ID,
		ChannelID: m.ChannelID,
		UserID:    m.UserID,
		Tier:      m.Tier,
		Status:    m.Status,
	}, nil
}

func (r *channelRepo) Subscribe(ctx context.Context, userID, channelID uint64) error {
	m := &model.Membership{
		ChannelID: channelID,
		UserID:    userID,
		Tier:      1,
		Status:    "active",
		StartedAt: time.Now(),
	}
	return r.data.DB.WithContext(ctx).Create(m).Error
}

func (r *channelRepo) Unsubscribe(ctx context.Context, userID, channelID uint64) error {
	return r.data.DB.WithContext(ctx).
		Where("user_id = ? AND channel_id = ?", userID, channelID).
		Delete(&model.Membership{}).Error
}

// HasMembership implements biz.MembershipChecker for VideoService.
// Given a viewer userID and the channel owner's userID, returns the membership tier.
func (r *channelRepo) HasMembership(ctx context.Context, userID, channelOwnerUserID uint64) (int8, error) {
	var ch model.Channel
	if err := r.data.DB.WithContext(ctx).Where("user_id = ?", channelOwnerUserID).First(&ch).Error; err != nil {
		return 0, err
	}
	var m model.Membership
	if err := r.data.DB.WithContext(ctx).
		Where("user_id = ? AND channel_id = ? AND status = ?", userID, ch.ID, "active").
		First(&m).Error; err != nil {
		return 0, err
	}
	return m.Tier, nil
}

func toBizChannel(m *model.Channel) *biz.Channel {
	displayName := ""
	avatarURL := ""
	if m.User.ID != 0 {
		displayName = m.User.DisplayName
		if m.User.AvatarURL != nil {
			avatarURL = *m.User.AvatarURL
		}
	}
	return &biz.Channel{
		ID:          m.ID,
		UserID:      m.UserID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		MonthlyFee:  m.MonthlyFee,
		IsHidden:    m.IsHidden,
	}
}
