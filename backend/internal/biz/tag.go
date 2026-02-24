package biz

import (
	"context"
	"math/rand"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type Tag struct {
	ID   uint64
	Name string
	Slug string
}

type TagRepo interface {
	ListAll(ctx context.Context) ([]*Tag, error)
	GetUserTags(ctx context.Context, userID *uint64, sessionID *string) ([]*Tag, error)
	SetUserTags(ctx context.Context, userID *uint64, sessionID *string, tagIDs []uint64) error
	GetTagsByIDs(ctx context.Context, ids []uint64) ([]*Tag, error)
}

type TagUsecase struct {
	repo TagRepo
	log  *log.Helper
}

func NewTagUsecase(repo TagRepo, logger log.Logger) *TagUsecase {
	return &TagUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *TagUsecase) ListTags(ctx context.Context) ([]*Tag, error) {
	return uc.repo.ListAll(ctx)
}

func (uc *TagUsecase) GetMyTags(ctx context.Context, userID *uint64, sessionID *string) ([]*Tag, error) {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return nil, nil // no preferences, return empty
	}
	return uc.repo.GetUserTags(ctx, userID, sessionID)
}

func (uc *TagUsecase) SetMyTags(ctx context.Context, userID *uint64, sessionID *string, tagIDs []uint64) ([]*Tag, error) {
	if userID == nil && (sessionID == nil || *sessionID == "") {
		return nil, errors.BadRequest("UNAUTHORIZED", "must be logged in or provide session_id")
	}

	if len(tagIDs) > 5 {
		return nil, errors.BadRequest("TAG_LIMIT_EXCEEDED", "maximum 5 tags allowed")
	}

	if len(tagIDs) == 0 {
		// Clear preferences
		if err := uc.repo.SetUserTags(ctx, userID, sessionID, nil); err != nil {
			return nil, err
		}
		return nil, nil
	}

	// Verify all tag IDs exist
	tags, err := uc.repo.GetTagsByIDs(ctx, tagIDs)
	if err != nil {
		return nil, err
	}
	if len(tags) != len(tagIDs) {
		return nil, errors.NotFound("TAG_NOT_FOUND", "one or more tag IDs not found")
	}

	if err := uc.repo.SetUserTags(ctx, userID, sessionID, tagIDs); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetRecommendedTagIDs returns a random subset of the user's tag preferences
// for recommendation variety. Each request gets a different combination.
func (uc *TagUsecase) GetRecommendedTagIDs(ctx context.Context, userID *uint64, sessionID *string) ([]uint64, error) {
	tags, err := uc.GetMyTags(ctx, userID, sessionID)
	if err != nil || len(tags) == 0 {
		return nil, err
	}

	// Randomly pick 1 to len(tags) tags for variety
	n := rand.Intn(len(tags)) + 1
	rand.Shuffle(len(tags), func(i, j int) { tags[i], tags[j] = tags[j], tags[i] })

	ids := make([]uint64, n)
	for i := 0; i < n; i++ {
		ids[i] = tags[i].ID
	}
	return ids, nil
}
