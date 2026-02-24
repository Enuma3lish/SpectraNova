package pagination

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

func Normalize(page, pageSize int32) (offset, limit int) {
	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	offset = int((page - 1) * pageSize)
	limit = int(pageSize)
	return
}
