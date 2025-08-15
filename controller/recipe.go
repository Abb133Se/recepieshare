package controller

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/service"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TopRatedRecipe struct {
	RecipeID   uint    `json:"recipe_id"`
	Title      string  `json:"title"`
	Average    float64 `joson:"average"`
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
	CommenterName string `json:"user_name"`
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	err = db.Preload("Ingredients").
		Preload("Comments").
		Preload("User").
		Preload("Tags").
		Preload("Categories").
		Preload("Steps").
		First(&recipe, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipe"})
		return
	}

	imageIDs, _ := service.GetImageIDsForEntity("recipe", recipe.ID)

	var favoriteCount int64
	db.Model(&model.Favorite{}).Where("recipe_id = ?", recipe.ID).Count(&favoriteCount)

	isFavorited := false
	if userIDVal, exists := c.Get("userID"); exists {
		if userID, ok := userIDVal.(uint); ok {
			var favExists int64
			db.Model(&model.Favorite{}).
				Where("recipe_id = ? AND user_id = ?", recipe.ID, userID).
				Count(&favExists)
			isFavorited = favExists > 0
		}
	}

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
		FavoriteCount: favoriteCount,
		IsFavorited:   isFavorited,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create recipe"})
		return
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
// @Description  Delete a recipe given its ID
// @Tags         recipes
// @Param        id   path      int  true  "Recipe ID"
// @Success      200  {object}  SimpleMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id} [delete]
func DeleteRecipeHandler(c *gin.Context) {
	var recepie model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "internal error",
		})
		return
	}

	err = db.First(&recepie, validID).Error
	if err != nil {
		c.JSON(404, gin.H{
			"message": "not found",
		})
		return
	}

	if err := db.Delete(&recepie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete recipe"})
		return
	}

	c.JSON(200, SimpleMessageResponse{
		Message: "Recipe deleted successfully",
	})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.First(&model.Recipe{}, validID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe does not exist"})
		return
	}

	err = db.Model(&model.Ingredient{}).Where("recipe_id = ?", c.Param("id")).Find(&Ingredient).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve Ingredient from server"})
		return
	}

	if len(Ingredient) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "recipe does not have Ingredient"})
		return
	}

	c.JSON(http.StatusOK, IngredientsResponse{
		Message: "Ingredient successfully retrieved",
		Data:    Ingredient,
	})

}

// GetAllRecipeCommentsHandler godoc
// @Summary      Get paginated comments for a recipe
// @Description  Retrieve comments with pagination and sorting for a recipe
// @Tags         recipes
// @Param        id      path      int     true  "Recipe ID"
// @Param        limit   query     int     false "Limit number of comments"
// @Param        offset  query     int     false "Offset for pagination"
// @Param        sort    query     string  false "Sort order (e.g., date_desc)"
// @Success      200     {object}  CommentsResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /recipe/{id}/comments [get]
func GetAllRecipeCommentsHandler(c *gin.Context) {
	var comments []model.Comment

	var limit, offset = 1, 0

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	sortParam := c.DefaultQuery("sort", "date_desc")

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.First(&model.Recipe{}, validID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	query := db.Model(&model.Comment{}).Where("recipe_id = ?", validID).Limit(limit).Offset(offset)
	query = utils.ApplyCommentSorting(query, sortParam)

	err = query.Find(&comments).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comments"})
		return
	}

	var commentCount int64
	err = db.Model(&model.Comment{}).Where("recipe_id = ?", validID).Count(&commentCount).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to count comments"})
		return
	}

	var responseComments []CommentWithUserName
	for _, comment := range comments {
		var user model.User
		var commenterName string
		if comment.UserID != 0 {
			if err := db.First(&user, comment.UserID).Error; err == nil {
				commenterName = user.Name
			}
		}
		responseComments = append(responseComments, CommentWithUserName{
			Comment:       comment,
			CommenterName: commenterName,
		})
	}

	c.JSON(http.StatusOK, CommentsResponse{
		Message: "comments retrieved successfully",
		Data:    responseComments,
		Count:   commentCount,
	})

}

// GetAllRecipesHandler godoc
// @Summary      Get paginated list of recipes
// @Description  Retrieve recipes with pagination, including tags, categories, steps, and image IDs
// @Tags         recipes
// @Param        limit   query  int  false "Limit number of recipes"
// @Param        offset  query  int  false "Offset for pagination"
// @Success      200     {object}  RecipeListWithImagesResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /recipe/list [get]
func GetAllRecipesHandler(c *gin.Context) {
	var recipes []model.Recipe

	limit, offset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to DB"})
		return
	}

	if err := db.Preload("Tags").Preload("Categories").
		Limit(limit).Offset(offset).Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipes"})
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
		Message: "recipes retrieved successfully",
		Data:    recipesWithImages,
	})
}

// PutRecipeUpdateHandler godoc
// @Summary      Update a recipe
// @Description  Updates the title, text, ingredients, and steps of a specific recipe
// @Tags         recipes
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "Recipe ID"
// @Param        recipe  body    object  true  "Updated recipe data"
// @Success      200     {object}  SimpleMessageResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /recipe/{id} [put]
func PutRecipeUpdateHandler(c *gin.Context) {
	var input struct {
		Title      string             `json:"title"`
		Text       string             `json:"text"`
		Ingredient []model.Ingredient `json:"ingridients"`
		Steps      []model.Step       `json:"steps"`
	}

	var recipe model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "faild to connect to db"})
		return
	}

	if err = db.First(&recipe, validID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipe"})
		return
	}

	recipe.Title = input.Title
	recipe.Text = input.Text

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&recipe).Error; err != nil {
			return err
		}

		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&model.Ingredient{}).Error; err != nil {
			return err
		}

		for _, ing := range input.Ingredient {
			ing.RecipeID = recipe.ID
			if err := tx.Create(&ing).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&model.Step{}).Error; err != nil {
			return err
		}

		for _, step := range input.Steps {
			step.RecipeID = recipe.ID
			if err := tx.Create(&step).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't update recipe", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "recipe updated successfully"})

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch top rated recipes"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "faild to fetch popular recipes"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	var ingredients []model.Ingredient
	if err := db.Where("recipe_id = ?", recipeID).Find(&ingredients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ingredients"})
		return
	}

	var ingredientStrings []string
	for _, ing := range ingredients {
		ingredientStrings = append(ingredientStrings, fmt.Sprintf("%s of %s", ing.Amount, ing.Name))
	}

	nutritionData, err := utils.EstimateNutrition(API_KEY, ingredientStrings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI model inference failed", "details": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	var recipe model.Recipe
	err = db.Preload("Tags").First(&recipe, recipeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve recipe tags"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Tags").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
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
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tag: " + tagName})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tag: " + tagName})
				return
			}
		}
		tags = append(tags, tag)
	}

	if err := db.Model(&recipe).Association("Tags").Replace(tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update recipe tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe tags updated", "tags": tags})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection failed"})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Tags").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	if err := db.Model(&recipe).Association("Tags").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear tags"})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "tags removed from recipe"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	var recipe model.Recipe
	err = db.Preload("Categories").First(&recipe, recipeID).Limit(limit).Offset(offset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve recipe categories"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection failed"})
		return
	}

	var recipe model.Recipe
	if err := db.Preload("Categories").First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}

	if err := db.Model(&recipe).Association("Categories").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear categories"})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "categories removed from recipe"})
}

// SearchRecipesHandler godoc
// @Summary      Search recipes
// @Description  Retrieve a list of recipes matching the given filters, including tags, categories, steps, and image IDs
// @Tags         recipes
// @Param        title         query   string  false  "Filter by recipe title (partial match)"
// @Param        ingredient    query   string  false  "Filter by ingredient name"
// @Param        tag_ids       query   string  false  "Comma-separated list of tag IDs"
// @Param        category_ids  query   string  false  "Comma-separated list of category IDs"
// @Param        user_id       query   string  false  "Filter by recipe author's user ID"
// @Param        sort          query   string  false  "Sort field (e.g., 'title', 'created_at')"
// @Param        limit         query   int     false  "Max number of recipes to return"
// @Param        offset        query   int     false  "Number of recipes to skip"
// @Success      200  {object}  RecipeListWithImagesResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipes/search [get]
func SearchRecipesHandler(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	params := map[string]string{
		"title":        c.Query("title"),
		"ingredient":   c.Query("ingredient"),
		"tag_ids":      c.Query("tag_ids"),
		"category_ids": c.Query("category_ids"),
		"user_id":      c.Query("user_id"),
	}

	query := db.Model(&model.Recipe{}).
		Preload("Tags").
		Preload("Categories").
		Preload("User")

	query = utils.ApplyRecipeFilters(query, params)
	query = utils.ApplyRecipeSorting(query, c.Query("sort"))

	limit, offset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var recipes []model.Recipe
	if err := query.Limit(limit).Offset(offset).Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
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
		Message: "recipes retrieved successfully",
		Data:    recipesWithImages,
	})
}
