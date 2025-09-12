package utils

import (
	"fmt"
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

	if minCalStr, ok := params["min_calories"]; ok && minCalStr != "" {
		if minCal, err := strconv.ParseFloat(minCalStr, 64); err == nil {
			query = query.Where("recipes.calories >= ?", minCal)
		}
	}
	if maxCalStr, ok := params["max_calories"]; ok && maxCalStr != "" {
		if maxCal, err := strconv.ParseFloat(maxCalStr, 64); err == nil {
			fmt.Println("SQL:", query.Statement.SQL.String())
			query = query.Where("recipes.calories <= ?", maxCal)
		}
	}
	if minProteinStr, ok := params["min_protein"]; ok && minProteinStr != "" {
		if minProtein, err := strconv.ParseFloat(minProteinStr, 64); err == nil {
			query = query.Where("recipes.protein >= ?", minProtein)
		}
	}
	if maxProteinStr, ok := params["max_protein"]; ok && maxProteinStr != "" {
		if maxProtein, err := strconv.ParseFloat(maxProteinStr, 64); err == nil {
			query = query.Where("recipes.protein <= ?", maxProtein)
		}
	}
	if minFatStr, ok := params["min_fat"]; ok && minFatStr != "" {
		if minFat, err := strconv.ParseFloat(minFatStr, 64); err == nil {
			query = query.Where("recipes.fat >= ?", minFat)
		}
	}
	if maxFatStr, ok := params["max_fat"]; ok && maxFatStr != "" {
		if maxFat, err := strconv.ParseFloat(maxFatStr, 64); err == nil {
			query = query.Where("recipes.fat <= ?", maxFat)
		}
	}
	if minCarbsStr, ok := params["min_carbs"]; ok && minCarbsStr != "" {
		if minCarbs, err := strconv.ParseFloat(minCarbsStr, 64); err == nil {
			query = query.Where("recipes.carbs >= ?", minCarbs)
		}
	}
	if maxCarbsStr, ok := params["max_carbs"]; ok && maxCarbsStr != "" {
		if maxCarbs, err := strconv.ParseFloat(maxCarbsStr, 64); err == nil {
			query = query.Where("recipes.carbs <= ?", maxCarbs)
		}
	}
	if minFiberStr, ok := params["min_fiber"]; ok && minFiberStr != "" {
		if minFiber, err := strconv.ParseFloat(minFiberStr, 64); err == nil {
			query = query.Where("recipes.fiber >= ?", minFiber)
		}
	}
	if maxFiberStr, ok := params["max_fiber"]; ok && maxFiberStr != "" {
		if maxFiber, err := strconv.ParseFloat(maxFiberStr, 64); err == nil {
			query = query.Where("recipes.fiber <= ?", maxFiber)
		}
	}
	if minSugarStr, ok := params["min_sugar"]; ok && minSugarStr != "" {
		if minSugar, err := strconv.ParseFloat(minSugarStr, 64); err == nil {
			query = query.Where("recipes.sugar >= ?", minSugar)
		}
	}
	if maxSugarStr, ok := params["max_sugar"]; ok && maxSugarStr != "" {
		if maxSugar, err := strconv.ParseFloat(maxSugarStr, 64); err == nil {
			query = query.Where("recipes.sugar <= ?", maxSugar)
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
	case "calories_asc":
		return query.Order("recipes.calories ASC")
	case "calories_desc":
		return query.Order("recipes.calories DESC")
	case "protein_asc":
		return query.Order("recipes.protein ASC")
	case "protein_desc":
		return query.Order("recipes.protein DESC")
	case "fat_asc":
		return query.Order("recipes.fat ASC")
	case "fat_desc":
		return query.Order("recipes.fat DESC")
	case "carbs_asc":
		return query.Order("recipes.carbs ASC")
	case "carbs_desc":
		return query.Order("recipes.carbs DESC")
	case "fiber_asc":
		return query.Order("recipes.fiber ASC")
	case "fiber_desc":
		return query.Order("recipes.fiber DESC")
	case "sugar_asc":
		return query.Order("recipes.sugar ASC")
	case "sugar_desc":
		return query.Order("recipes.sugar DESC")
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

func ApplyCatSorting(query *gorm.DB, sortParam string) *gorm.DB {
	switch strings.ToLower(sortParam) {
	case "title_desc":
		return query.Order("categories.name DESC")
	case "title_asc":
		return query.Order("categories.name ASC")
	case "date_asc":
		return query.Order("categories.created_at ASC")
	case "date_desc":
		return query.Order("categories.created_at DESC")
	default:
		return query.Order("categories.created_at DESC") // default sorting
	}
}

func GetTableAndColumn(metric string) (string, string, error) {
	switch metric {
	case "views":
		return "recipe_views", "created_at", nil
	case "favorites":
		return "favorites", "created_at", nil
	case "ratings":
		return "ratings", "created_at", nil
	case "site":
		return "site_visits", "created_at", nil
	case "recipes":
		return "recipes", "created_at", nil
	default:
		return "", "", fmt.Errorf("invalid metric")
	}
}
