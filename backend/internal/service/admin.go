package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
	"backend/internal/pkg/authctx"
)

type AdminService struct {
	v1.UnimplementedAdminServiceServer
	uc *biz.AdminUsecase
}

func NewAdminService(uc *biz.AdminUsecase) *AdminService {
	return &AdminService{uc: uc}
}

func (s *AdminService) AdminListUsers(ctx context.Context, req *v1.AdminListUsersRequest) (*v1.AdminListUsersReply, error) {
	users, total, err := s.uc.ListUsers(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.AdminUserInfo, len(users))
	for i, u := range users {
		items[i] = &v1.AdminUserInfo{
			Id:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Role:        u.Role,
			IsHidden:    u.IsHidden,
			CreatedAt:   u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &v1.AdminListUsersReply{
		Users: items,
		Total: total,
	}, nil
}

func (s *AdminService) AdminDeleteUser(ctx context.Context, req *v1.AdminDeleteUserRequest) (*v1.AdminDeleteUserReply, error) {
	callerID, _ := authctx.UserIDFromContext(ctx)
	if err := s.uc.DeleteUser(ctx, callerID, req.Id); err != nil {
		return nil, err
	}
	return &v1.AdminDeleteUserReply{}, nil
}

func (s *AdminService) AdminListVideos(ctx context.Context, req *v1.AdminListVideosRequest) (*v1.AdminListVideosReply, error) {
	videos, total, err := s.uc.ListAllVideos(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.AdminVideoInfo, len(videos))
	for i, v := range videos {
		items[i] = &v1.AdminVideoInfo{
			Id:             v.ID,
			Title:          v.Title,
			Username:       v.Username,
			UserId:         v.UserID,
			CategoryName:   v.CategoryName,
			AccessTier:     int32(v.AccessTier),
			IsPublished:    v.IsPublished,
			IsHidden:       v.IsHidden,
			ViewsMember:    v.ViewsMember,
			ViewsNonMember: v.ViewsNonMember,
			CreatedAt:      v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &v1.AdminListVideosReply{
		Videos: items,
		Total:  total,
	}, nil
}

func (s *AdminService) AdminDeleteVideo(ctx context.Context, req *v1.AdminDeleteVideoRequest) (*v1.AdminDeleteVideoReply, error) {
	if err := s.uc.DeleteVideo(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.AdminDeleteVideoReply{}, nil
}

func (s *AdminService) AdminCreateTag(ctx context.Context, req *v1.AdminCreateTagRequest) (*v1.AdminCreateTagReply, error) {
	tag, err := s.uc.CreateTag(ctx, &biz.AdminTag{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		return nil, err
	}
	return &v1.AdminCreateTagReply{
		Tag: &v1.AdminTagInfo{Id: tag.ID, Name: tag.Name, Slug: tag.Slug},
	}, nil
}

func (s *AdminService) AdminUpdateTag(ctx context.Context, req *v1.AdminUpdateTagRequest) (*v1.AdminUpdateTagReply, error) {
	tag, err := s.uc.UpdateTag(ctx, &biz.AdminTag{
		ID:   req.Id,
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		return nil, err
	}
	return &v1.AdminUpdateTagReply{
		Tag: &v1.AdminTagInfo{Id: tag.ID, Name: tag.Name, Slug: tag.Slug},
	}, nil
}

func (s *AdminService) AdminDeleteTag(ctx context.Context, req *v1.AdminDeleteTagRequest) (*v1.AdminDeleteTagReply, error) {
	if err := s.uc.DeleteTag(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.AdminDeleteTagReply{}, nil
}
