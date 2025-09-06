package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/service"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TopRatedRecipe struct {
	RecipeID   uint    `json:"recipe_id"`
	Title      string  `json:"title"`
	Average    float64 `json:"average"`
	TotalVotes int64   `json:"total_votes"`
}

type MostPopularRecipe struct {
	RecipeID      uint   `json:"recipe_id"`
	Title         string `json:"title"`
	FavoriteCount int64  `json:"favorite_count"`
}

type TagNamesInput struct {
	Tags []string `json:"tags" binding:"required"`
}

type PostRecipeRequest struct {
	Title       string             `json:"title" binding:"required"`
	Text        string             `json:"text" binding:"required"`
	Ingredients []model.Ingredient `json:"ingredients" binding:"required"`
	TagIDs      []uint             `json:"tag_ids"`   // Optional: Use existing tag IDs
	TagNames    []string           `json:"tag_names"` // Optional: Create/find tags by name
	CategoryIDs []uint             `json:"category_ids"`
	Steps       []model.Step       `json:"steps"`
}

type RecipeResponse struct {
	Message string       `json:"message"`
	Data    model.Recipe `json:"data"`
}

type RecipeListResponse struct {
	Message string         `json:"message"`
	Data    []model.Recipe `json:"data"`
}

type SimpleMessageResponse struct {
	Message string `json:"message"`
}

type TopRatedRecipesResponse []TopRatedRecipe
type MostPopularRecipesResponse struct {
	Recipes []MostPopularRecipe `json:"recipes"`
}

type IngredientsResponse struct {
	Message string             `json:"message"`
	Data    []model.Ingredient `json:"data"`
}

type CommentsResponse struct {
	Message string                `json:"message"`
	Data    []CommentWithUserName `json:"data"`
	Count   int64                 `json:"count"`
}

type CommentWithUserName struct {
	model.Comment
	UserName string `json:"user_name"`
}

type TagsResponse struct {
	Tags []model.Tag `json:"tags"`
}

type RecipeCategoriesResponse struct {
	Categories []model.Category `json:"categories"`
}

type NutritionResponse struct {
	NutritionalValues interface{} `json:"nutritional_values"`
}

type RecipeWithImageIDs struct {
	ID            uint               `json:"id"`
	Title         string             `json:"title"`
	Text          string             `json:"text"`
	UserID        uint               `json:"user_id"`
	Ingredients   []model.Ingredient `json:"ingredients"`
	Tags          []model.Tag        `json:"tags"`
	Categories    []model.Category   `json:"categories"`
	Steps         []model.Step       `json:"steps"`
	Images        []uint             `json:"images"`
	FavoriteCount int64              `json:"favorite_count"`
	IsFavorited   bool               `json:"is_favorited"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type RecipeListWithImagesResponse struct {
	Message string               `json:"message"`
	Data    []RecipeWithImageIDs `json:"data"`
	Count   int64                `json:"count"`
}

var API_KEY = "YTAMecQ6C06ClaR/HmS26g==OUlc0LkiJgLyFjhv"

// GetRecipeHandler godoc
// @Summary      Get recipe by ID
// @Description  Get detailed recipe info including ingredients, comments, tags, categories, steps, and image IDs
// @Tags         recipes
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  RecipeWithImageIDs
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id} [get]
func GetRecipeHandler(c *gin.Context) {
	var recipe model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	// Preload associations
	err = db.Preload("Ingredients").
		Preload("Comments").
		Preload("User").
		Preload("Tags").
		Preload("Categories").
		Preload("Steps").
		First(&recipe, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	imageIDs, _ := service.GetImageIDsForEntity("recipe", recipe.ID)

	type FavResult struct {
		Count       int64
		UserFavored bool
	}

	var result FavResult
	userID := uint(0)
	if uidVal, exists := c.Get("userID"); exists {
		if uid, ok := uidVal.(uint); ok {
			userID = uid
		}
	}

	db.Model(&model.Favorite{}).
		Select("COUNT(*) as count, SUM(CASE WHEN user_id = ? THEN 1 ELSE 0 END) > 0 as user_favored", userID).
		Where("recipe_id = ?", recipe.ID).
		Scan(&result)

	resp := RecipeWithImageIDs{
		ID:            recipe.ID,
		Title:         recipe.Title,
		Text:          recipe.Text,
		UserID:        recipe.UserID,
		Ingredients:   recipe.Ingredients,
		Tags:          recipe.Tags,
		Categories:    recipe.Categories,
		Steps:         recipe.Steps,
		Images:        imageIDs,
		FavoriteCount: result.Count,
		IsFavorited:   result.UserFavored,
		CreatedAt:     recipe.CreatedAt,
		UpdatedAt:     recipe.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// PostRecipeHandler godoc
// @Summary      Create a new recipe
// @Description  Create a new recipe with ingredients, tags (by IDs or names), categories, steps, and returns the new recipe including image IDs
// @Tags         recipes
// @Accept       json
// @Produce      json
// @Param        recipe  body      PostRecipeRequest  true  "Recipe data"
// @Success      201     {object}  RecipeWithImageIDs
// @Failure      400     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /recipe [post]
func PostRecipeHandler(c *gin.Context) {
	var req PostRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.User.UserNotFound.String()})
		return
	}

	var tags []model.Tag
	if len(req.TagIDs) > 0 {
		db.Find(&tags, req.TagIDs)
	}
	for _, name := range req.TagNames {
		var tag model.Tag
		if err := db.Where("name = ?", name).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = model.Tag{Name: name}
				db.Create(&tag)
			}
		}
		tags = append(tags, tag)
	}

	var categories []model.Category
	if len(req.CategoryIDs) > 0 {
		db.Find(&categories, req.CategoryIDs)
	}

	recipe := model.Recipe{
		Title:       req.Title,
		Text:        req.Text,
		UserID:      userID,
		Ingredients: req.Ingredients,
		Tags:        tags,
		Categories:  categories,
		Steps:       req.Steps,
	}

	if err := db.Create(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeCreateFailed.String()})
		return
	}

	var ingredientStrings []string
	for _, ing := range recipe.Ingredients {
		ingredientStrings = append(ingredientStrings, fmt.Sprintf("%s of %s", ing.Amount, ing.Name))
	}

	nutritionItems, err := utils.EstimateNutrition(API_KEY, ingredientStrings)
	if err == nil && len(nutritionItems) == len(recipe.Ingredients) {
		var totalCalories, totalProtein, totalFat, totalCarbs, totalFiber, totalSugar float64

		for i, item := range nutritionItems {
			recipe.Ingredients[i].Calories = item.Calories
			recipe.Ingredients[i].Protein = item.ProteinG
			recipe.Ingredients[i].Fat = item.FatTotalG
			recipe.Ingredients[i].Carbs = item.CarbohydratesTotalG
			recipe.Ingredients[i].Fiber = item.FiberG
			recipe.Ingredients[i].Sugar = item.SugarG

			totalCalories += item.Calories
			totalProtein += item.ProteinG
			totalFat += item.FatTotalG
			totalCarbs += item.CarbohydratesTotalG
			totalFiber += item.FiberG
			totalSugar += item.SugarG
		}

		recipe.Calories = totalCalories
		recipe.Protein = totalProtein
		recipe.Fat = totalFat
		recipe.Carbs = totalCarbs
		recipe.Fiber = totalFiber
		recipe.Sugar = totalSugar

		db.Save(&recipe)
		for i := range recipe.Ingredients {
			db.Save(&recipe.Ingredients[i])
		}
	}

	imageIDs, _ := service.GetImageIDsForEntity("recipe", recipe.ID)

	resp := RecipeWithImageIDs{
		ID:          recipe.ID,
		Title:       recipe.Title,
		Text:        recipe.Text,
		Ingredients: req.Ingredients,
		UserID:      recipe.UserID,
		Tags:        recipe.Tags,
		Categories:  recipe.Categories,
		Steps:       recipe.Steps,
		Images:      imageIDs,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   recipe.UpdatedAt,
	}

	c.JSON(http.StatusCreated, resp)
}

// DeleteRecipeHandler godoc
// @Summary      Delete a recipe by ID
// @Description  Delete a recipe and all its associated entities if it belongs to the authenticated user
// @Tags         recipes
// @Security     BearerAuth
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  SimpleMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id} [delete]

// DeleteRecipeHandler (admin route) godoc
// @Summary      Delete a recipe (admin)
// @Description  Admin deletes a recipe of any user
// @Tags         recipes
// @Security     BearerAuth
// @Param        userID path int true "User ID"
// @Param        id     path int true "Recipe ID"
// @Success      200 {object} controller.SimpleMessageResponse
// @Failure      400 {object} controller.ErrorResponse
// @Failure      401 {object} controller.ErrorResponse
// @Failure      403 {object} controller.ErrorResponse
// @Failure      404 {object} controller.ErrorResponse
// @Failure      500 {object} controller.ErrorResponse
// @Router       /admin/user/{userID}/recipe/{id} [delete]
func DeleteRecipeHandler(c *gin.Context) {
	var userID uint
	var err error
	if role := c.GetString("role"); role == "user" {
		userID := c.GetUint("userID")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
			return
		}
	} else if role == "admin" {
		userID, err = middleware.GetEffectiveUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
			return
		}
	}

	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.Where("id = ? AND user_id = ?", recipeID, userID).
		First(&recipe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: messages.Recipe.RecipeDeleteForbidden.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	// Clear M2M associations (join tables must be cleaned manually)
	_ = db.Model(&recipe).Association("Tags").Clear()
	_ = db.Model(&recipe).Association("Categories").Clear()

	// Delete recipe (cascade takes care of children)
	if err := db.Delete(&recipe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeDeleteFail.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.Recipe.RecipeDeleted.String()})
}

// GetAllRecipeIngredientHandler godoc
// @Summary      Get all ingredients for a recipe
// @Description  Retrieve all ingredients for a specific recipe by ID
// @Tags         recipes
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  IngredientsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/ingredients [get]
func GetAllRecipeIngredientHandler(c *gin.Context) {

	var Ingredient []model.Ingredient

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.First(&model.Recipe{}, validID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	err = db.Model(&model.Ingredient{}).Where("recipe_id = ?", c.Param("id")).Find(&Ingredient).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeIngredientsFetchFail.String()})
		return
	}

	if len(Ingredient) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeHasNoIngredient.String()})
		return
	}

	c.JSON(http.StatusOK, IngredientsResponse{
		Message: messages.Common.Success.String(),
		Data:    Ingredient,
	})

}

// GetAllRecipeCommentsHandler godoc
// @Summary      Get all comments for a recipe with pagination and sorting
// @Description  Retrieve a paginated list of comments for a specific recipe with total count, optionally sorted by likes or date
// @Tags         comments
// @Produce      json
// @Param        id    path      int     true   "Recipe ID"
// @Param        limit query     int     false  "Limit number of comments returned"
// @Param        offset query    int     false  "Number of comments to skip"
// @Param        sort  query     string  false  "Sort order: likes_desc, likes_asc, date_asc, date_desc"
// @Success      200   {object}  controller.CommentsResponse
// @Failure      400   {object}  controller.ErrorResponse
// @Failure      404   {object}  controller.ErrorResponse
// @Failure      500   {object}  controller.ErrorResponse
// @Router       /recipe/{id}/comments [get]
func GetAllRecipeCommentsHandler(c *gin.Context) {
	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.First(&recipe, validID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	sort := c.Query("sortOrder")
	query := db.Model(&model.Comment{}).
		Select("comments.*, users.name AS user_name").
		Joins("JOIN users ON users.id = comments.user_id").
		Where("comments.recipe_id = ?", validID)
	query = utils.ApplyCommentSorting(query, sort)

	// Count
	totalCount, err := utils.Count(query, "comments")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Comment.CommnetFetchFail.String()})
		return
	}

	// Paginate
	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}

	var comments []CommentWithUserName
	if err := utils.Paginate(query, limit, offset, &comments); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Comment.CommnetFetchFail.String()})
		return
	}

	c.JSON(http.StatusOK, CommentsResponse{
		Message: messages.Common.Success.String(),
		Data:    comments,
		Count:   totalCount,
	})
}

// GetAllRecipesHandler godoc
// @Summary      Get all recipes with pagination, filtering, and sorting
// @Description  Retrieve a paginated list of recipes with total count, optionally filtered by title, ingredient, tags, categories, user, and sorted by title, creation date, rating, or favorites
// @Tags         recipes
// @Produce      json
// @Param        limit         query     int     false  "Limit number of recipes returned"
// @Param        offset        query     int     false  "Number of recipes to skip"
// @Param        sort          query     string  false  "Sort order: title_asc, title_desc, created_asc, created_desc, rating_desc, favorites_desc"
// @Param        title         query     string  false  "Filter by recipe title (partial match)"
// @Param        ingredient    query     string  false  "Filter by ingredient name (partial match)"
// @Param        tag_ids       query     string  false  "Filter by tag IDs (comma-separated)"
// @Param        category_ids  query     string  false  "Filter by category IDs (comma-separated)"
// @Param        user_id       query     int     false  "Filter by user ID"
// @Success      200           {object}  controller.RecipeListWithImagesResponse
// @Failure      400           {object}  controller.ErrorResponse
// @Failure      500           {object}  controller.ErrorResponse
// @Router       /recipe/list [get]
func GetAllRecipesHandler(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	params := map[string]string{
		"title":        c.Query("title"),
		"ingredient":   c.Query("ingredient"),
		"tag_ids":      c.Query("tag_ids"),
		"category_ids": c.Query("category_ids"),
		"user_id":      c.Query("user_id"),
		"rating":       c.Query("rating"),
		"min_calories": c.Query("min_calories"),
		"max_calories": c.Query("max_calories"),
		"min_protein":  c.Query("min_protein"),
		"max_protein":  c.Query("max_protein"),
		"min_fat":      c.Query("min_fat"),
		"max_fat":      c.Query("max_fat"),
		"min_carbs":    c.Query("min_carbs"),
		"max_carbs":    c.Query("max_carbs"),
		"min_fiber":    c.Query("min_fiber"),
		"max_fiber":    c.Query("max_fiber"),
		"min_sugar":    c.Query("min_sugar"),
		"max_sugar":    c.Query("max_sugar"),
	}
	query := db.Model(&model.Recipe{}).Preload("Ingredients").Preload("Tags").Preload("Categories").Preload("Steps")
	query = utils.ApplyRecipeFilters(query, params)
	sort := c.Query("sortOrder")

	query = utils.ApplyRecipeSorting(query, sort)

	var baseRecipes []model.Recipe
	totalCount, err := utils.Count(query, "recipes")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}
	if err := utils.Paginate(query, limit, offset, &baseRecipes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	var responseData []RecipeWithImageIDs
	for _, recipe := range baseRecipes {
		imageIDs, _ := service.GetImageIDsForEntity("recipe", recipe.ID)
		responseData = append(responseData, RecipeWithImageIDs{
			ID:          recipe.ID,
			Title:       recipe.Title,
			Text:        recipe.Text,
			UserID:      recipe.UserID,
			Ingredients: recipe.Ingredients,
			Tags:        recipe.Tags,
			Categories:  recipe.Categories,
			Steps:       recipe.Steps,
			Images:      imageIDs,
			CreatedAt:   recipe.CreatedAt,
			UpdatedAt:   recipe.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, RecipeListWithImagesResponse{
		Message: messages.Common.Success.String(),
		Data:    responseData,
		Count:   totalCount,
	})
}

// PutRecipeUpdateHandler godoc
// @Summary      Update an existing recipe
// @Description  Updates a recipe by ID. Replaces title, text, ingredients, steps, tags, and categories.
// @Tags         recipes
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Recipe ID"
// @Param        recipe body     object true "Updated recipe data"
// @Success      200   {object}  controller.SimpleMessageResponse
// @Failure      400   {object}  controller.SimpleMessageResponse
// @Failure      404   {object}  controller.SimpleMessageResponse
// @Failure      500   {object}  controller.SimpleMessageResponse
// @Router       /recipe/{id} [put]
func PutRecipeUpdateHandler(c *gin.Context) {
	var input struct {
		Title       string             `json:"title"`
		Text        string             `json:"text"`
		Ingredients []model.Ingredient `json:"ingredients"`
		Steps       []model.Step       `json:"steps"`
		TagIDs      []uint             `json:"tag_ids"`
		Tags        []model.Tag        `json:"tags"`
		CategoryIDs []uint             `json:"category_ids"`
	}

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, SimpleMessageResponse{Message: err.Error()})
		return
	}

	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, SimpleMessageResponse{Message: "invalid input: " + err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err = db.Preload("Ingredients").Preload("Steps").
		Preload("Tags").Preload("Categories").
		First(&recipe, validID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	recipe.Title = input.Title
	recipe.Text = input.Text

	err = db.Transaction(func(tx *gorm.DB) error {
		// Save recipe base fields
		if err := tx.Save(&recipe).Error; err != nil {
			return err
		}

		// Replace ingredients
		for i := range input.Ingredients {
			input.Ingredients[i].RecipeID = recipe.ID
		}
		if err := tx.Model(&recipe).Association("Ingredients").Replace(input.Ingredients); err != nil {
			return err
		}

		// Replace steps
		for i := range input.Steps {
			input.Steps[i].RecipeID = recipe.ID
		}
		if err := tx.Model(&recipe).Association("Steps").Replace(input.Steps); err != nil {
			return err
		}

		// Handle tags (deduplicate and create new if necessary)
		tagMap := make(map[uint]model.Tag)
		tagNameMap := make(map[string]bool)

		if len(input.TagIDs) > 0 {
			var existingTags []model.Tag
			if err := tx.Where("id IN ?", input.TagIDs).Find(&existingTags).Error; err != nil {
				return err
			}
			for _, t := range existingTags {
				tagMap[t.ID] = t
				tagNameMap[strings.ToLower(t.Name)] = true
			}
		}

		for _, t := range input.Tags {
			tagName := strings.ToLower(strings.TrimSpace(t.Name))
			if tagName == "" || tagNameMap[tagName] {
				continue
			}

			var existing model.Tag
			if err := tx.Where("LOWER(name) = ?", tagName).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := tx.Create(&t).Error; err != nil {
						return err
					}
					tagMap[t.ID] = t
					tagNameMap[tagName] = true
				} else {
					return err
				}
			} else {
				tagMap[existing.ID] = existing
				tagNameMap[tagName] = true
			}
		}

		var finalTags []model.Tag
		for _, t := range tagMap {
			finalTags = append(finalTags, t)
		}

		if err := tx.Model(&recipe).Association("Tags").Replace(finalTags); err != nil {
			return err
		}

		// Handle categories
		var categories []model.Category
		if len(input.CategoryIDs) > 0 {
			if err := tx.Where("id IN ?", input.CategoryIDs).Find(&categories).Error; err != nil {
				return err
			}
			if len(categories) != len(input.CategoryIDs) {
				return fmt.Errorf("one or more category IDs are invalid")
			}
		}
		if err := tx.Model(&recipe).Association("Categories").Replace(categories); err != nil {
			return err
		}

		if err := tx.Preload("Ingredients").First(&recipe, recipe.ID).Error; err != nil {
			return err
		}

		var ingredientStrings []string
		for _, ing := range recipe.Ingredients {
			ingredientStrings = append(ingredientStrings, fmt.Sprintf("%s of %s", ing.Amount, ing.Name))
		}

		nutritionItems, err := utils.EstimateNutrition(API_KEY, ingredientStrings)
		if err == nil && len(nutritionItems) == len(recipe.Ingredients) {
			var totalCalories, totalProtein, totalFat, totalCarbs, totalFiber, totalSugar float64

			for i := range recipe.Ingredients {
				recipe.Ingredients[i].Calories = nutritionItems[i].Calories
				recipe.Ingredients[i].Protein = nutritionItems[i].ProteinG
				recipe.Ingredients[i].Fat = nutritionItems[i].FatTotalG
				recipe.Ingredients[i].Carbs = nutritionItems[i].CarbohydratesTotalG
				recipe.Ingredients[i].Fiber = nutritionItems[i].FiberG
				recipe.Ingredients[i].Sugar = nutritionItems[i].SugarG

				totalCalories += nutritionItems[i].Calories
				totalProtein += nutritionItems[i].ProteinG
				totalFat += nutritionItems[i].FatTotalG
				totalCarbs += nutritionItems[i].CarbohydratesTotalG
				totalFiber += nutritionItems[i].FiberG
				totalSugar += nutritionItems[i].SugarG
			}

			recipe.Calories = totalCalories
			recipe.Protein = totalProtein
			recipe.Fat = totalFat
			recipe.Carbs = totalCarbs
			recipe.Fiber = totalFiber
			recipe.Sugar = totalSugar

			if err := tx.Save(&recipe).Error; err != nil {
				return err
			}

			for i := range recipe.Ingredients {
				if err := tx.Save(&recipe.Ingredients[i]).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeUpdateFail.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.Recipe.RecipeUpdated.String()})
}

// GetTopRatedRecipesHandler godoc
// @Summary      Get top rated recipes
// @Description  Get recipes sorted by average rating, with total votes count
// @Tags         recipes
// @Param        limit   query  int  false "Limit number of recipes"
// @Param        offset  query  int  false "Offset for pagination"
// @Success      200     {array}  TopRatedRecipe
// @Failure      400     {object} ErrorResponse
// @Failure      500     {object} ErrorResponse
// @Router       /recipes/top-rated [get]
func GetTopRatedRecipesHandler(c *gin.Context) {
	var results []TopRatedRecipe
	var limit, offset = 1, 0

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.Table("ratings").
		Select("recipes.id AS recipe_id, recipes.title, AVG(ratings.score) AS average, COUNT(ratings.id) AS total_votes").
		Joins("JOIN recipes ON recipes.id = ratings.recipe_id").
		Group("ratings.recipe_id").
		Order("average DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	c.JSON(http.StatusOK, results)

}

// GetMostPopularRecipesHandler godoc
// @Summary      Get most popular recipes by favorites
// @Description  Get recipes sorted by number of favorites
// @Tags         recipes
// @Param        limit   query  int  false "Limit number of recipes"
// @Param        offset  query  int  false "Offset for pagination"
// @Success      200     {object} MostPopularRecipesResponse
// @Failure      400     {object} ErrorResponse
// @Failure      500     {object} ErrorResponse
// @Router       /recipes/most-popular [get]
func GetMostPopularRecipesHandler(c *gin.Context) {
	var results []MostPopularRecipe
	var limit, offset = 1, 0

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.Table("recipes").
		Select("recipes.id as recipe_id, recipes.title, COUNT(favorites.id) as favorite_count").
		Joins("LEFT JOIN favorites ON recipes.id = favorites.recipe_id").
		Group("recipes.id").
		Order("favorite_count DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	c.JSON(http.StatusOK, MostPopularRecipesResponse{Recipes: results})
}

// GetRecipeNutritionHandler godoc
// @Summary      Get nutritional values for a recipe
// @Description  Estimate nutrition info based on ingredients via AI model
// @Tags         recipes
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  NutritionResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/calories [get]
func GetRecipeNutritionHandler(c *gin.Context) {
	id := c.Param("id")

	recipeID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var ingredients []model.Ingredient
	if err := db.Where("recipe_id = ?", recipeID).Find(&ingredients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeIngredientsFetchFail.String()})
		return
	}

	var ingredientStrings []string
	for _, ing := range ingredients {
		ingredientStrings = append(ingredientStrings, fmt.Sprintf("%s of %s", ing.Amount, ing.Name))
	}

	nutritionData, err := utils.EstimateNutrition(API_KEY, ingredientStrings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeNutritionFail.String()})
		return
	}

	c.JSON(http.StatusOK, NutritionResponse{NutritionalValues: nutritionData})
}

// GetRecipeTagsHandler godoc
// @Summary      Get tags for a recipe
// @Description  Retrieves all tags associated with the specified recipe
// @Tags         recipes
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  TagsResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/tags [get]
func GetRecipeTagsHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	err = db.Preload("Tags").First(&recipe, recipeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeTagFetchFail.String()})
		}
		return
	}

	c.JSON(http.StatusOK, TagsResponse{Tags: recipe.Tags})
}

// PutRecipeTagsHandler godoc
// @Summary      Update tags for a recipe
// @Description  Replaces the tags associated with the specified recipe.
//
//	If a tag does not exist, it will be created.
//
// @Tags         recipes
// @Accept       json
// @Produce      json
// @Param        id    path   int            true  "Recipe ID"
// @Param        tags  body   TagNamesInput  true  "List of tag names"
// @Success      200   {object}  TagsResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /recipe/{id}/tags [put]
func PutRecipeTagsHandler(c *gin.Context) {
	recipeID := c.Param("id")

	_, err := utils.ValidateEntityID(recipeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Tags").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	var input TagNamesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	var tags []model.Tag
	for _, tagName := range input.Tags {
		var tag model.Tag
		if err := db.Where("name = ?", tagName).First(&tag).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				tag = model.Tag{Name: tagName}
				if err := db.Create(&tag).Error; err != nil {
					c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeTagCreateFail.String()})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeTagQueryFail.String()})
				return
			}
		}
		tags = append(tags, tag)
	}

	if err := db.Model(&recipe).Association("Tags").Replace(tags); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeTagUpdateFail.String()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Recipe.RecipeTagUpdated.String(), "tags": tags})
}

// DeleteRecipeTagsHandler godoc
// @Summary      Remove all tags from a recipe
// @Description  Clears all tags associated with the specified recipe.
// @Tags         recipes
// @Produce      json
// @Param        id   path   int  true  "Recipe ID"
// @Success      200  {object}  SimpleMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/tags [delete]
func DeleteRecipeTagsHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Tags").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	if err := db.Model(&recipe).Association("Tags").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeTagsDeleteFail.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.Recipe.ReciepTagsDeleted.String()})
}

// GetRecipeCategoriesHandler godoc
// @Summary      Get categories for a recipe
// @Description  Retrieves all categories associated with the specified recipe.
// @Tags         recipes
// @Produce      json
// @Param        id   path   int  true  "Recipe ID"
// @Success      200  {object}  CategoriesResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/categories [get]
func GetRecipeCategoriesHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	var limit, offset = 1, 0
	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	err = db.Preload("Categories").First(&recipe, recipeID).Limit(limit).Offset(offset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeCatsFetcFail.String()})
		}
		return
	}

	c.JSON(http.StatusOK, RecipeCategoriesResponse{Categories: recipe.Categories})
}

// DeleteRecipeCategoriesHandler godoc
// @Summary      Remove all categories from a recipe
// @Description  Clears all categories associated with the specified recipe.
// @Tags         recipes
// @Produce      json
// @Param        id   path   int  true  "Recipe ID"
// @Success      200  {object}  SimpleMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/categories [delete]
func DeleteRecipeCategoriesHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Categories").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	if err := db.Model(&recipe).Association("Categories").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeCatsDeleteFail.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.Recipe.RecipeCatsDeleted.String()})
}

// SearchRecipesHandler godoc
// @Summary      Search recipes with pagination, filtering, and sorting
// @Description  Search for recipes using various filters and retrieve a paginated list with total count, optionally sorted
// @Tags         recipes
// @Produce      json
// @Param        title         query     string  false  "Filter by recipe title (partial match)"
// @Param        ingredient    query     string  false  "Filter by ingredient name (partial match)"
// @Param        tag_ids       query     string  false  "Filter by tag IDs (comma-separated)"
// @Param        category_ids  query     string  false  "Filter by category IDs (comma-separated)"
// @Param        user_id       query     string  false  "Filter by user ID"
// @Param        sort          query     string  false  "Sort order: title_asc, title_desc, created_asc, created_desc, rating_desc, favorites_desc"
// @Param        limit         query     int     false  "Limit number of recipes returned"
// @Param        offset        query     int     false  "Number of recipes to skip"
// @Success      200           {object}  controller.RecipeListWithImagesResponse
// @Failure      400           {object}  controller.ErrorResponse
// @Failure      500           {object}  controller.ErrorResponse
// @Router       /recipes/search [get]
func SearchRecipesHandler(c *gin.Context) {
	fmt.Println("Full query string:", c.Request.URL.RawQuery)
	params := map[string]string{
		"title":        c.Query("title"),
		"ingredient":   c.Query("ingredient"),
		"tag_ids":      c.Query("tag_ids"),
		"category_ids": c.Query("category_ids"),
		"user_id":      c.Query("user_id"),
		"rating":       c.Query("rating"),
		"min_calories": c.Query("min_calories"),
		"max_calories": c.Query("max_calories"),
		"min_protein":  c.Query("min_protein"),
		"max_protein":  c.Query("max_protein"),
		"min_fat":      c.Query("min_fat"),
		"max_fat":      c.Query("max_fat"),
		"min_carbs":    c.Query("min_carbs"),
		"max_carbs":    c.Query("max_carbs"),
		"min_fiber":    c.Query("min_fiber"),
		"max_fiber":    c.Query("max_fiber"),
		"min_sugar":    c.Query("min_sugar"),
		"max_sugar":    c.Query("max_sugar"),
	}
	fmt.Println(c.Query("max_calories"))

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	query := db.Model(&model.Recipe{}).
		Preload("Tags").
		Preload("Categories").
		Preload("User").
		Preload("Ratings").
		Preload("Ingredients")
	query = utils.ApplyRecipeFilters(query, params)
	query = utils.ApplyRecipeSorting(query, c.Query("sortOrder"))

	var recipes []model.Recipe
	totalCount, err := utils.Count(query, "recipes")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}
	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}
	if err := utils.Paginate(query, limit, offset, &recipes); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	var recipesWithImages []RecipeWithImageIDs
	for _, r := range recipes {
		imageIDs, _ := service.GetImageIDsForEntity("recipe", r.ID)
		recipesWithImages = append(recipesWithImages, RecipeWithImageIDs{
			ID:         r.ID,
			Title:      r.Title,
			Text:       r.Text,
			UserID:     r.UserID,
			Tags:       r.Tags,
			Categories: r.Categories,
			Steps:      r.Steps,
			Images:     imageIDs,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, RecipeListWithImagesResponse{
		Message: messages.Common.Success.String(),
		Data:    recipesWithImages,
		Count:   totalCount,
	})
}
