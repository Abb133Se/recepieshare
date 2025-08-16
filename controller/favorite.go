package controller

import (
	"net/http"
	"strconv"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	// Check if recipe exists
	if err := db.Select("id").First(&model.Recipe{}, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "recipe not found"})
		return
	}

	// Insert favorite only if not exists
	favorite := model.Favorite{UserID: userID, RecipeID: req.RecipeID}
	if err := db.FirstOrCreate(&favorite, model.Favorite{UserID: userID, RecipeID: req.RecipeID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to add favorite"})
		return
	}

	c.JSON(http.StatusOK, FavoriteResponse{
		Message:    "favorite added successfully",
		FavoriteID: favorite.ID,
	})
}

// DeleteFavoriteHandler godoc
// @Summary      Remove a recipe from favorites
// @Description  Unfavorites a recipe for the authenticated user using its recipe ID
// @Tags         favorites
// @Produce      json
// @Param        recipeId  path      int  true  "Recipe ID"
// @Success      200   {object}  SimpleMessageResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /unfavorite [delete]
func DeleteFavoriteHandler(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	// Get logged-in user ID from context (set by your auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
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
			c.JSON(http.StatusNotFound, gin.H{"error": "favorite not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query favorite"})
		}
		return
	}

	// Delete the favorite
	if err := db.Delete(&favorite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe unfavorited successfully"})

}
