package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"
	"backend/internal/pkg/pagination"

	"github.com/go-kratos/kratos/v2/log"
)

type searchRepo struct {
	data *Data
	log  *log.Helper
}

func NewSearchRepo(data *Data, logger log.Logger) biz.SearchRepo {
	return &searchRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *searchRepo) Search(ctx context.Context, params *biz.SearchParams) ([]*biz.Video, int64, error) {
	query := r.data.DB.WithContext(ctx).
		Model(&model.Video{}).
		Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false)

	// FULLTEXT search on title (BOOLEAN MODE for small datasets)
	if params.Query != "" {
		query = query.Where("MATCH(videos.title) AGAINST(? IN BOOLEAN MODE)", params.Query)
	}

	// Filters
	if params.CategoryID != nil {
		query = query.Where("videos.category_id = ?", *params.CategoryID)
	}
	if params.MinDuration != nil {
		query = query.Where("videos.duration >= ?", *params.MinDuration)
	}
	if params.MaxDuration != nil {
		query = query.Where("videos.duration <= ?", *params.MaxDuration)
	}
	if params.DateFrom != nil {
		query = query.Where("videos.created_at >= ?", *params.DateFrom)
	}
	if params.DateTo != nil {
		query = query.Where("videos.created_at <= ?", *params.DateTo)
	}
	if params.AccessType != "" {
		switch params.AccessType {
		case "public":
			query = query.Where("videos.access_tier = 0")
		case "member":
			query = query.Where("videos.access_tier > 0")
		}
	}

	// Count before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	switch params.SortBy {
	case "views_desc":
		query = query.Order("(videos.views_member + videos.views_non_member) DESC")
	case "views_asc":
		query = query.Order("(videos.views_member + videos.views_non_member) ASC")
	case "date_asc":
		query = query.Order("videos.created_at ASC")
	default:
		query = query.Order("videos.created_at DESC")
	}

	// Pagination
	offset, limit := pagination.Normalize(params.Page, params.PageSize)

	var videos []model.Video
	if err := query.
		Preload("Tags").Preload("Category").Preload("User").
		Offset(offset).Limit(limit).
		Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return toBizVideos(videos), total, nil
}
