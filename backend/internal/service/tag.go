package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
	"backend/internal/pkg/authctx"
)

type TagService struct {
	v1.UnimplementedTagServiceServer
	uc *biz.TagUsecase
}

func NewTagService(uc *biz.TagUsecase) *TagService {
	return &TagService{uc: uc}
}

func (s *TagService) ListTags(ctx context.Context, req *v1.ListTagsRequest) (*v1.TagListReply, error) {
	tags, err := s.uc.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.TagListReply{Tags: toBizTagItems(tags)}, nil
}

func (s *TagService) GetMyTags(ctx context.Context, req *v1.GetMyTagsRequest) (*v1.TagListReply, error) {
	userID, sessionID := extractTagIdentity(ctx, req.SessionId)

	tags, err := s.uc.GetMyTags(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	return &v1.TagListReply{Tags: toBizTagItems(tags)}, nil
}

func (s *TagService) SetMyTags(ctx context.Context, req *v1.SetMyTagsRequest) (*v1.TagListReply, error) {
	userID, sessionID := extractTagIdentity(ctx, req.SessionId)

	tags, err := s.uc.SetMyTags(ctx, userID, sessionID, req.TagIds)
	if err != nil {
		return nil, err
	}
	return &v1.TagListReply{Tags: toBizTagItems(tags)}, nil
}

// extractTagIdentity gets user_id from JWT context or session_id from request.
func extractTagIdentity(ctx context.Context, sessionID *string) (*uint64, *string) {
	uid, ok := authctx.UserIDFromContext(ctx)
	if ok {
		return &uid, nil
	}
	return nil, sessionID
}

func toBizTagItems(tags []*biz.Tag) []*v1.TagItem {
	if tags == nil {
		return nil
	}
	items := make([]*v1.TagItem, len(tags))
	for i, t := range tags {
		items[i] = &v1.TagItem{
			Id:   t.ID,
			Name: t.Name,
			Slug: t.Slug,
		}
	}
	return items
}
