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
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}

	if ingredient, ok := params["ingredient"]; ok && ingredient != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM ingredients WHERE ingredients.recipe_id = recipes.id AND LOWER(ingredients.name) LIKE ?)",
			"%"+strings.ToLower(ingredient)+"%",
		)
	}

	if tagIDsStr, ok := params["tag_ids"]; ok && tagIDsStr != "" {
		tagIDs := ParseUintSlice(tagIDsStr)
		if len(tagIDs) > 0 {
			query = query.Where(
				"EXISTS (SELECT 1 FROM recipe_tags WHERE recipe_tags.recipe_id = recipes.id AND recipe_tags.tag_id IN ?)",
				tagIDs,
			)
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

	// Ensure unique recipes when joins are applied
	return query.Select("recipes.*").Group("recipes.id")

}

func ApplyRecipeSorting(query *gorm.DB, sortParam string) *gorm.DB {
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
		return query.Select("recipes.*").
			Joins("LEFT JOIN ratings ON ratings.recipe_id = recipes.id").
			Group("recipes.id").
			Order("AVG(ratings.score) DESC")
	case "favorites_desc":
		return query.Select("recipes.*").
			Joins("LEFT JOIN favorites ON favorites.recipe_id = recipes.id").
			Group("recipes.id").
			Order("COUNT(favorites.id) DESC")
	default:
		return query.Order("created_at DESC")
	}
}

func ApplyCommentSorting(query *gorm.DB, sortParam string) *gorm.DB {
	switch strings.ToLower(sortParam) {
	case "likes_desc":
		return query.Order("likes DESC")
	case "likes_asc":
		return query.Order("likes ASC")
	case "date_asc":
		return query.Order("created_at ASC")
	case "date_desc":
		return query.Order("created_at DESC")
	default:
		return query.Order("created_at DESC") // default sorting
	}
}
