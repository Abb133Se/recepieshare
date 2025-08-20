package utils

import (
	"gorm.io/gorm"
)

func Count(query *gorm.DB, primaryTable string) (int64, error) {
	var total int64

	countQuery := query.Session(&gorm.Session{})

	// Use qualified id for counting
	countQuery = countQuery.Select(primaryTable + ".id")

	// Remove ORDER BY for counting
	delete(countQuery.Statement.Clauses, "ORDER BY")

	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func Paginate[T any](query *gorm.DB, limit, offset int, result *[]T) error {
	return query.Limit(limit).Offset(offset * limit).Find(result).Error
}
