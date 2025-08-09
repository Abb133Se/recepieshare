package controller

import (
	"net/http"
	"strconv"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	var favorite model.Favorite
	var user model.User
	var recipe model.Recipe

	err := c.BindJSON(&favorite)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	_, err = utils.ValidateEntityID(strconv.Itoa(int(favorite.ID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = utils.ValidateEntityID(strconv.Itoa(int(favorite.RecipeID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = utils.ValidateEntityID(strconv.Itoa(int(favorite.UserID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	if err = db.First(&recipe, favorite.RecipeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipe data"})
		return
	}
	if err = db.First(&user, favorite.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user data"})
		return
	}

	err = db.Create(&favorite).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add favorite"})
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
