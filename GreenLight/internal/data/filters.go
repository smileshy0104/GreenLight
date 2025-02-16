package data

import (
	"DesignMode/GreenLight/internal/validator"
	"math"
	"strings"
)

// Filters 用于封装分页信息
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

// Metadata 包含分页信息的结构体
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// calculateMetadata 函数用于计算分页信息
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{} // return an empty Metadata struct if there are no records
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// ValidateFilters 校验过滤器
func ValidateFilters(v *validator.Validator, f Filters) {
	// 检查page和page_size参数
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_0000, "", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	// 检查sort参数
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

// sortColumn 排序字段
func (f Filters) sortColumn() string {
	// 遍历排序字段，如果存在则返回排序字段，否则抛出panic
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// 如果排序字段不存在，抛出panic
	panic("unsafe sort parameter:" + f.Sort)
}

// sortDirection 排序方向
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// limit 分页条数
func (f Filters) limit() int {
	return f.PageSize
}

// offset 分页偏移量
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}
