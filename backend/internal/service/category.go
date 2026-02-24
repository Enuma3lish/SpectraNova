package service

import (
	"context"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/biz"
)

type CategoryService struct {
	v1.UnimplementedCategoryServiceServer
	uc *biz.CategoryUsecase
}

func NewCategoryService(uc *biz.CategoryUsecase) *CategoryService {
	return &CategoryService{uc: uc}
}

func (s *CategoryService) ListCategories(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesReply, error) {
	categories, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CategoryItem, len(categories))
	for i, c := range categories {
		items[i] = &v1.CategoryItem{
			Id:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		}
	}
	return &v1.ListCategoriesReply{Categories: items}, nil
}
