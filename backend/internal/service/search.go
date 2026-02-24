package service

import (
	"context"
	"time"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
)

type SearchService struct {
	v1.UnimplementedSearchServiceServer
	uc *biz.SearchUsecase
}

func NewSearchService(uc *biz.SearchUsecase) *SearchService {
	return &SearchService{uc: uc}
}

func (s *SearchService) Search(ctx context.Context, req *v1.SearchRequest) (*v1.VideoListReply, error) {
	params := &biz.SearchParams{
		Query:    req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.CategoryId != nil {
		params.CategoryID = req.CategoryId
	}
	if req.MinDuration != nil {
		params.MinDuration = req.MinDuration
	}
	if req.MaxDuration != nil {
		params.MaxDuration = req.MaxDuration
	}
	if req.DateFrom != nil {
		if t, err := time.Parse("2006-01-02", *req.DateFrom); err == nil {
			params.DateFrom = &t
		}
	}
	if req.DateTo != nil {
		if t, err := time.Parse("2006-01-02", *req.DateTo); err == nil {
			params.DateTo = &t
		}
	}
	if req.SortBy != nil {
		params.SortBy = *req.SortBy
	}
	if req.AccessType != nil {
		params.AccessType = *req.AccessType
	}

	videos, total, err := s.uc.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]*v1.VideoReply, len(videos))
	for i, v := range videos {
		items[i] = toVideoReply(v)
	}
	return &v1.VideoListReply{Videos: items, Total: total}, nil
}
