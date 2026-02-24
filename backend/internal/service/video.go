package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
	"backend/internal/pkg/authctx"

	"github.com/go-kratos/kratos/v2/errors"
)

type VideoService struct {
	v1.UnimplementedVideoServiceServer
	uc *biz.VideoUsecase
}

func NewVideoService(uc *biz.VideoUsecase) *VideoService {
	return &VideoService{uc: uc}
}

func (s *VideoService) CreateVideo(ctx context.Context, req *v1.CreateVideoRequest) (*v1.VideoReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	desc := ""
	if req.Description != nil {
		desc = *req.Description
	}
	thumb := ""
	if req.ThumbnailUrl != nil {
		thumb = *req.ThumbnailUrl
	}

	tags := make([]*biz.Tag, len(req.TagIds))
	for i, id := range req.TagIds {
		tags[i] = &biz.Tag{ID: id}
	}

	video, err := s.uc.CreateVideo(ctx, userID, &biz.Video{
		Title:        req.Title,
		Description:  desc,
		CategoryID:   req.CategoryId,
		VideoURL:     req.VideoUrl,
		ThumbnailURL: thumb,
		Duration:     req.Duration,
		AccessTier:   int8(req.AccessTier),
		Tags:         tags,
	})
	if err != nil {
		return nil, err
	}
	return toVideoReply(video), nil
}

func (s *VideoService) GetVideo(ctx context.Context, req *v1.GetVideoRequest) (*v1.VideoReply, error) {
	var viewerID *uint64
	uid, ok := authctx.UserIDFromContext(ctx)
	if ok {
		viewerID = &uid
	}
	role, _ := authctx.RoleFromContext(ctx)

	video, err := s.uc.GetVideo(ctx, req.Id, viewerID, role)
	if err != nil {
		return nil, err
	}
	return toVideoReply(video), nil
}

func (s *VideoService) UpdateVideo(ctx context.Context, req *v1.UpdateVideoRequest) (*v1.VideoReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	v := &biz.Video{ID: req.Id}
	if req.Title != nil {
		v.Title = *req.Title
	}
	if req.Description != nil {
		v.Description = *req.Description
	}
	if req.CategoryId != nil {
		v.CategoryID = *req.CategoryId
	}
	if req.ThumbnailUrl != nil {
		v.ThumbnailURL = *req.ThumbnailUrl
	}
	if req.AccessTier != nil {
		v.AccessTier = int8(*req.AccessTier)
	}
	if len(req.TagIds) > 0 {
		tags := make([]*biz.Tag, len(req.TagIds))
		for i, id := range req.TagIds {
			tags[i] = &biz.Tag{ID: id}
		}
		v.Tags = tags
	}

	video, err := s.uc.UpdateVideo(ctx, userID, v)
	if err != nil {
		return nil, err
	}
	return toVideoReply(video), nil
}

func (s *VideoService) DeleteVideo(ctx context.Context, req *v1.DeleteVideoRequest) (*v1.DeleteVideoReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	if err := s.uc.DeleteVideo(ctx, userID, req.Id); err != nil {
		return nil, err
	}
	return &v1.DeleteVideoReply{Success: true}, nil
}

func (s *VideoService) TogglePublish(ctx context.Context, req *v1.TogglePublishRequest) (*v1.VideoReply, error) {
	userID, ok := authctx.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("UNAUTHORIZED", "login required")
	}

	video, err := s.uc.TogglePublish(ctx, userID, req.Id, req.IsPublished)
	if err != nil {
		return nil, err
	}
	return toVideoReply(video), nil
}

func (s *VideoService) GetRecommended(ctx context.Context, req *v1.GetRecommendedRequest) (*v1.VideoListReply, error) {
	var userID *uint64
	uid, ok := authctx.UserIDFromContext(ctx)
	if ok {
		userID = &uid
	}

	videos, total, err := s.uc.GetRecommended(ctx, userID, req.SessionId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.VideoReply, len(videos))
	for i, v := range videos {
		items[i] = toVideoReply(v)
	}
	return &v1.VideoListReply{Videos: items, Total: total}, nil
}

func toVideoReply(v *biz.Video) *v1.VideoReply {
	if v == nil {
		return nil
	}
	tags := make([]*v1.TagItem, len(v.Tags))
	for i, t := range v.Tags {
		tags[i] = &v1.TagItem{Id: t.ID, Name: t.Name, Slug: t.Slug}
	}
	return &v1.VideoReply{
		Id:           v.ID,
		UserId:       v.UserID,
		Username:     v.Username,
		CategoryId:   v.CategoryID,
		CategoryName: v.CategoryName,
		Title:        v.Title,
		Description:  v.Description,
		VideoUrl:     v.VideoURL,
		ThumbnailUrl: v.ThumbnailURL,
		Duration:     v.Duration,
		Views:        v.ViewsMember + v.ViewsNonMember,
		AccessTier:   int32(v.AccessTier),
		IsPublished:  v.IsPublished,
		Tags:         tags,
		CreatedAt:    v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
