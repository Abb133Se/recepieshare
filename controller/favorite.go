package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
)

type PostFavoriteRequest struct {
	RecipeID uint `json:"recipe_id" binding:"required"`
}

type FavoriteResponse struct {
	Message    string `json:"message"`
	FavoriteID uint   `json:"id"`
}

// PostFavoriteHandler godoc
// @Summary      Add a favorite recipe for a user
// @Description  Adds a favorite record linking a user and a recipe
// @Tags         favorites
// @Accept       json
// @Produce      json
// @Param        favorite  body      model.Favorite  true  "Favorite data"
// @Success      200       {object}  FavoriteResponse
// @Failure      400       {object}  ErrorResponse
// @Failure      404       {object}  ErrorResponse
// @Failure      500       {object}  ErrorResponse
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
// @Summary      Delete a favorite by ID
// @Description  Deletes a favorite record by its ID
// @Tags         favorites
// @Produce      json
// @Param        id    path      int  true  "Favorite ID"
// @Success      200   {object}  SimpleMessageResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Router       /favorite/{id} [delete]
func DeleteFavoriteHandler(c *gin.Context) {
	var favorite model.Favorite

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fialed to connect to db"})
		return
	}

	if err = db.First(&favorite, validID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "favorite does not exists"})
		return
	}

	err = db.Delete(&model.Favorite{}, validID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete favorite"})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "favorite deleted successfully"})

}
