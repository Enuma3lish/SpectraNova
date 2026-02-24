package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type SearchParams struct {
	Query       string
	CategoryID  *uint64
	MinDuration *uint32
	MaxDuration *uint32
	DateFrom    *time.Time
	DateTo      *time.Time
	SortBy      string
	AccessType  string
	Page        int32
	PageSize    int32
}

type SearchRepo interface {
	Search(ctx context.Context, params *SearchParams) ([]*Video, int64, error)
}

type SearchUsecase struct {
	repo SearchRepo
	log  *log.Helper
}

func NewSearchUsecase(repo SearchRepo, logger log.Logger) *SearchUsecase {
	return &SearchUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *SearchUsecase) Search(ctx context.Context, params *SearchParams) ([]*Video, int64, error) {
	return uc.repo.Search(ctx, params)
}
