package utils

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func ParseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []uint
	for _, p := range parts {
		if val, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64); err == nil {
			result = append(result, uint(val))
		}
	}
	return result
}

func ApplyRecipeFilters(query *gorm.DB, params map[string]string) *gorm.DB {
	if title, ok := params["title"]; ok && title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	if ingredient, ok := params["ingredient"]; ok && ingredient != "" {
		query = query.Joins("JOIN ingredients ON ingredients.recipe_id = recipes.id").
			Where("ingredients.name LIKE ?", "%"+ingredient+"%")
	}

	if tagIDsStr, ok := params["tag_ids"]; ok && tagIDsStr != "" {
		tagIDs := ParseUintSlice(tagIDsStr)
		if len(tagIDs) > 0 {
			query = query.Joins("JOIN recipe_tags ON recipe_tags.recipe_id = recipes.id").
				Where("recipe_tags.tag_id IN ?", tagIDs)
		}
	}

	if categoryIDsStr, ok := params["category_ids"]; ok && categoryIDsStr != "" {
		categoryIDs := ParseUintSlice(categoryIDsStr)
		if len(categoryIDs) > 0 {
			query = query.Joins("JOIN recipe_categories ON recipe_categories.recipe_id = recipes.id").
				Where("recipe_categories.category_id IN ?", categoryIDs)
		}
	}

	if userIDStr, ok := params["user_id"]; ok && userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			query = query.Where("user_id = ?", userID)
		}
	}

	return query
}

func ApplySorting(query *gorm.DB, sortParam string) *gorm.DB {
	switch sortParam {
	case "title_asc":
		return query.Order("title ASC")
	case "title_desc":
		return query.Order("title DESC")
	case "created_asc":
		return query.Order("created_at ASC")
	case "created_desc":
		return query.Order("created_at DESC")
	case "rating_desc":
		return query.Joins("LEFT JOIN ratings ON ratings.recipe_id = recipes.id").
			Group("recipes.id").
			Order("AVG(ratings.score) DESC")
	case "favorites_desc":
		return query.Joins("LEFT JOIN favorites ON favorites.recipe_id = recipes.id").
			Group("recipes.id").
			Order("COUNT(favorites.id) DESC")
	default:
		return query.Order("created_at DESC")
	}
}
