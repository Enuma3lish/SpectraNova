package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
	"backend/internal/pkg/authctx"

	"github.com/go-kratos/kratos/v2/errors"
)

type ChannelService struct {
	v1.UnimplementedChannelServiceServer
	uc *biz.ChannelUsecase
}

func NewChannelService(uc *biz.ChannelUsecase) *ChannelService {
	return &ChannelService{uc: uc}
}

func (s *ChannelService) GetChannel(ctx context.Context, req *v1.GetChannelRequest) (*v1.ChannelReply, error) {
	var viewerID *uint64
	uid, ok := authctx.UserIDFromContext(ctx)
	if ok {
		viewerID = &uid
	}

	ch, membershipStatus, err := s.uc.GetChannel(ctx, req.Id, viewerID)
	if err != nil {
		return nil, err
	}
	return &v1.ChannelReply{
		Id:               ch.ID,
		UserId:           ch.UserID,
		DisplayName:      ch.DisplayName,
		AvatarUrl:        ch.AvatarURL,
		MonthlyFee:       ch.MonthlyFee,
		SubscriberCount:  ch.SubscriberCount,
		MembershipStatus: membershipStatus,
	}, nil
}

func (s *ChannelService) Subscribe(ctx context.Context, req *v1.SubscribeRequest) (*v1.MembershipReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	if err := s.uc.Subscribe(ctx, userID, req.Id); err != nil {
		return nil, err
	}
	return &v1.MembershipReply{
		Status: "subscribed",
		Tier:   "free",
	}, nil
}

func (s *ChannelService) Unsubscribe(ctx context.Context, req *v1.UnsubscribeRequest) (*v1.UnsubscribeReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	if err := s.uc.Unsubscribe(ctx, userID, req.Id); err != nil {
		return nil, err
	}
	return &v1.UnsubscribeReply{Status: "unsubscribed"}, nil
}
