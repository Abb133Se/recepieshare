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
		query = query.Where("LOWER(recipes.title) LIKE ?", "%"+strings.ToLower(title)+"%")
	}

	if ingredient, ok := params["ingredient"]; ok && ingredient != "" {
		query = query.Where(
			"EXISTS (SELECT 1 FROM ingredients i WHERE i.recipe_id = recipes.id AND LOWER(i.name) LIKE ?)",
			"%"+strings.ToLower(ingredient)+"%",
		)
	}

	if tagIDsStr, ok := params["tag_ids"]; ok && tagIDsStr != "" {
		tagIDs := ParseUintSlice(tagIDsStr)
		if len(tagIDs) > 0 {
			query = query.Where(
				"EXISTS (SELECT 1 FROM recipe_tags rt WHERE rt.recipe_id = recipes.id AND rt.tag_id IN ?)",
				tagIDs,
			)
		}
	}

	if categoryIDsStr, ok := params["category_ids"]; ok && categoryIDsStr != "" {
		categoryIDs := ParseUintSlice(categoryIDsStr)
		if len(categoryIDs) > 0 {
			query = query.Where(
				"EXISTS (SELECT 1 FROM recipe_categories rc WHERE rc.recipe_id = recipes.id AND rc.category_id IN ?)",
				categoryIDs,
			)
		}
	}

	if userIDStr, ok := params["user_id"]; ok && userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
			query = query.Where("recipes.user_id = ?", userID)
		}
	}

	return query
}

func ApplyRecipeSorting(query *gorm.DB, sortParam string) *gorm.DB {
	switch sortParam {
	case "title_asc":
		return query.Order("recipes.title ASC")
	case "title_desc":
		return query.Order("recipes.title DESC")
	case "created_asc":
		return query.Order("recipes.created_at ASC")
	case "created_desc":
		return query.Order("recipes.created_at DESC")
	case "rating_desc":
		return query.
			Joins("LEFT JOIN ratings r ON r.recipe_id = recipes.id").
			Group("recipes.id").
			Order("AVG(r.score) DESC")
	case "favorites_desc":
		return query.
			Joins("LEFT JOIN favorites f ON f.recipe_id = recipes.id").
			Group("recipes.id").
			Order("COUNT(f.id) DESC")
	default:
		return query.Order("recipes.created_at DESC")
	}
}

func ApplyCommentSorting(query *gorm.DB, sortParam string) *gorm.DB {
	switch strings.ToLower(sortParam) {
	case "likes_desc":
		return query.Order("comments.likes DESC")
	case "likes_asc":
		return query.Order("comments.likes ASC")
	case "date_asc":
		return query.Order("comments.created_at ASC")
	case "date_desc":
		return query.Order("comments.created_at DESC")
	default:
		return query.Order("comments.created_at DESC") // default sorting
	}
}
