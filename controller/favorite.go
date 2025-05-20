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

	c.JSON(http.StatusOK, gin.H{
		"message": "favorite added successfully",
		"id":      favorite.ID,
	})
}

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

	c.JSON(http.StatusOK, gin.H{"message": "favorite deleted successfully"})

}
