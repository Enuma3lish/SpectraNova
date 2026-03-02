package biz

import (
	"context"
	"time"

	"backend/internal/pkg/pagination"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type AdminUser struct {
	ID          uint64
	Username    string
	DisplayName string
	Role        string
	IsHidden    bool
	CreatedAt   time.Time
}

type AdminVideo struct {
	ID             uint64
	Title          string
	Username       string
	UserID         uint64
	CategoryName   string
	AccessTier     int8
	IsPublished    bool
	IsHidden       bool
	ViewsMember    uint64
	ViewsNonMember uint64
	CreatedAt      time.Time
}

type AdminTag struct {
	ID   uint64
	Name string
	Slug string
}

type AdminRepo interface {
	ListUsers(ctx context.Context, offset, limit int) ([]*AdminUser, int64, error)
	FindUserByID(ctx context.Context, id uint64) (*AdminUser, error)
	DeleteUser(ctx context.Context, id uint64) error
	ListAllVideos(ctx context.Context, offset, limit int) ([]*AdminVideo, int64, error)
	DeleteVideo(ctx context.Context, id uint64) error
	CreateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error)
	UpdateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error)
	DeleteTag(ctx context.Context, id uint64) error
	FindTagByID(ctx context.Context, id uint64) (*AdminTag, error)
	FindTagByName(ctx context.Context, name string) (*AdminTag, error)
}

type AdminUsecase struct {
	repo AdminRepo
	log  *log.Helper
}

func NewAdminUsecase(repo AdminRepo, logger log.Logger) *AdminUsecase {
	return &AdminUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *AdminUsecase) ListUsers(ctx context.Context, page, pageSize int32) ([]*AdminUser, int64, error) {
	offset, limit := pagination.Normalize(page, pageSize)
	return uc.repo.ListUsers(ctx, offset, limit)
}

func (uc *AdminUsecase) DeleteUser(ctx context.Context, callerID, targetID uint64) error {
	if callerID == targetID {
		return errors.BadRequest("ADMIN_SELF_DELETE", "admin cannot delete their own account")
	}

	_, err := uc.repo.FindUserByID(ctx, targetID)
	if err != nil {
		return errors.NotFound("USER_NOT_FOUND", "user not found")
	}

	if err := uc.repo.DeleteUser(ctx, targetID); err != nil {
		return errors.InternalServer("INTERNAL", "failed to delete user")
	}

	return nil
}

func (uc *AdminUsecase) ListAllVideos(ctx context.Context, page, pageSize int32) ([]*AdminVideo, int64, error) {
	offset, limit := pagination.Normalize(page, pageSize)
	return uc.repo.ListAllVideos(ctx, offset, limit)
}

func (uc *AdminUsecase) DeleteVideo(ctx context.Context, videoID uint64) error {
	if err := uc.repo.DeleteVideo(ctx, videoID); err != nil {
		return errors.InternalServer("INTERNAL", "failed to delete video")
	}
	return nil
}

func (uc *AdminUsecase) CreateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error) {
	existing, _ := uc.repo.FindTagByName(ctx, tag.Name)
	if existing != nil {
		return nil, errors.Conflict("TAG_ALREADY_EXISTS", "tag name already exists")
	}

	created, err := uc.repo.CreateTag(ctx, tag)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL", "failed to create tag")
	}
	return created, nil
}

func (uc *AdminUsecase) UpdateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error) {
	_, err := uc.repo.FindTagByID(ctx, tag.ID)
	if err != nil {
		return nil, errors.NotFound("TAG_NOT_FOUND", "tag not found")
	}

	// Check name uniqueness (exclude current tag)
	existing, _ := uc.repo.FindTagByName(ctx, tag.Name)
	if existing != nil && existing.ID != tag.ID {
		return nil, errors.Conflict("TAG_ALREADY_EXISTS", "tag name already exists")
	}

	updated, err := uc.repo.UpdateTag(ctx, tag)
	if err != nil {
		return nil, errors.InternalServer("INTERNAL", "failed to update tag")
	}
	return updated, nil
}

func (uc *AdminUsecase) DeleteTag(ctx context.Context, tagID uint64) error {
	_, err := uc.repo.FindTagByID(ctx, tagID)
	if err != nil {
		return errors.NotFound("TAG_NOT_FOUND", "tag not found")
	}

	if err := uc.repo.DeleteTag(ctx, tagID); err != nil {
		return errors.InternalServer("INTERNAL", "failed to delete tag")
	}
	return nil
}
