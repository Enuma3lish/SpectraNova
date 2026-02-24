package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewAuthUsecase,
	NewCategoryUsecase,
	NewTagUsecase,
	NewVideoUsecase,
	NewSearchUsecase,
	NewChannelUsecase,
)
