package controller

import (
	"net/http"
	"strconv"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostFavoriteRequest struct {
	RecipeID uint `json:"recipe_id" binding:"required"`
}

type FavoriteResponse struct {
	Message    string `json:"message"`
	FavoriteID uint   `json:"id"`
}

// PostFavoriteHandler godoc
// @Summary      Add a favorite recipe for the authenticated user
// @Description  Adds a favorite linking the current user (from JWT) and the specified recipe. A user can only favorite a recipe once.
// @Tags         favorites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        favorite  body      PostFavoriteRequest  true  "Favorite data"
// @Success      200       {object}  FavoriteResponse
// @Failure      400       {object}  ErrorResponse "Invalid request"
// @Failure      401       {object}  ErrorResponse "Unauthorized"
// @Failure      404       {object}  ErrorResponse "Recipe not found"
// @Failure      409       {object}  ErrorResponse "Favorite already exists"
// @Failure      500       {object}  ErrorResponse "Internal server error"
// @Router       /favorite [post]
func PostFavoriteHandler(c *gin.Context) {
	var req PostFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	// Check if recipe exists
	if err := db.Select("id").First(&model.Recipe{}, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	// Insert favorite only if not exists
	favorite := model.Favorite{UserID: userID, RecipeID: req.RecipeID}
	if err := db.FirstOrCreate(&favorite, model.Favorite{UserID: userID, RecipeID: req.RecipeID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Favorite.FavoriteAddFail.String()})
		return
	}

	c.JSON(http.StatusOK, FavoriteResponse{
		Message:    messages.Favorite.FavoriteAdded.String(),
		FavoriteID: favorite.ID,
	})
}

// DeleteFavoriteHandler godoc
// @Summary      Delete a favorite
// @Description  Deletes a favorite entry for the authenticated user
// @Tags         favorites
// @Security     BearerAuth
// @Param        recipe_id  query     int  true  "Recipe ID to remove from favorites"
// @Success      200  {object}  SuccessMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /favorite [delete]

// DeleteFavoriteHandler (admin route) godoc
// @Summary      Delete a favorite (admin)
// @Description  Admin removes a favorite of any user
// @Tags         favorites
// @Security     BearerAuth
// @Param        userID    path int true "User ID"
// @Param        favoriteID path int true "Favorite ID"
// @Success      200 {object} controller.SuccessMessageResponse
// @Failure      400 {object} controller.ErrorResponse
// @Failure      401 {object} controller.ErrorResponse
// @Failure      404 {object} controller.ErrorResponse
// @Failure      500 {object} controller.ErrorResponse
// @Router       /admin/user/{userID}/unfavorite/{favoriteID} [delete]
func DeleteFavoriteHandler(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	// Get logged-in user ID from context (set by your auth middleware)
	var userID uint
	if role := c.GetString("role"); role == "user" {
		userID = c.GetUint("userID")
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
	// Get recipeId from request (query or body depending on design)
	recipeIDParam := c.Query("recipe_id")
	if recipeIDParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipeId is required"})
		return
	}

	recipeID, err := strconv.Atoi(recipeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipeId"})
		return
	}

	// Find the favorite entry by userId + recipeId
	var favorite model.Favorite
	if err := db.Where("user_id = ? AND recipe_id = ?", userID, recipeID).First(&favorite).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Favorite.FavoriteNotFound.String()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Favorite.FavoriteRemoveQueryFail.String()})
		}
		return
	}

	// Delete the favorite
	if err := db.Delete(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Favorite.FavoriteRemoveFail.String()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Favorite.FavoriteRemoved.String()})

}

// GetAllFavorites godoc
// @Summary      Get all favorites
// @Description  Retrieve a paginated list of all favorites with user and recipe details
// @Tags         favorites
// @Security     BearerAuth
// @Param        limit     query     int     false  "Number of items per page" default(10)
// @Param        offset    query     int     false  "Pagination offset" default(0)
// @Param        sortOrder query     string  false  "Sort order: date_asc, date_desc" default(date_desc)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /favorites [get]
func GetAllFavorites(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}

	// Select favorites with user and recipe info
	query := db.Table("favorites").
		Select(`favorites.id as favorite_id, 
                users.id as user_id, 
                users.name as user_name,
                recipes.id as recipe_id, 
                recipes.title as recipe_title`).
		Joins("JOIN users ON favorites.user_id = users.id").
		Joins("JOIN recipes ON favorites.recipe_id = recipes.id")

	// Count total favorites
	total, err := utils.Count(query, "favorites")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count favorites"})
		return
	}

	// Fetch paginated results
	var favorites []struct {
		FavoriteID  uint   `json:"favorite_id"`
		UserID      uint   `json:"user_id"`
		UserName    string `json:"user_name"`
		RecipeID    uint   `json:"recipe_id"`
		RecipeTitle string `json:"recipe_title"`
	}

	if err := utils.Paginate(query, limit, offset, &favorites); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"favorites": favorites,
	})
}
