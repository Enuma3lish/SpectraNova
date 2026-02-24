package data

import (
	"context"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type categoryRepo struct {
	data *Data
	log  *log.Helper
}

func NewCategoryRepo(data *Data, logger log.Logger) biz.CategoryRepo {
	return &categoryRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *categoryRepo) ListAll(ctx context.Context) ([]*biz.Category, error) {
	var categories []model.Category
	if err := r.data.DB.WithContext(ctx).Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	result := make([]*biz.Category, len(categories))
	for i, c := range categories {
		result[i] = &biz.Category{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		}
	}
	return result, nil
}
