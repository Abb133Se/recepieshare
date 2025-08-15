package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PaginateAndCount[T any](c *gin.Context, query *gorm.DB, result *[]T) (int64, error) {
	limit, offset, err := ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil && !errors.Is(err, errors.New("invalid limit and offset values")) {
		return 0, err
	}
	if limit == 0 {
		limit = 10
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return 0, err
	}

	paginatedQuery := query.Limit(limit).Offset(offset)

	if err := paginatedQuery.Find(result).Error; err != nil {
		return 0, err
	}

	return totalCount, nil
}
