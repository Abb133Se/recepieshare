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

func PostRatingHandler(c *gin.Context) {
	var rating model.Rating
	var user model.User
	var recipe model.Recipe

	err := c.BindJSON(&rating)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	_, err = utils.ValidateEntityID(strconv.Itoa(int(rating.ID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = utils.ValidateEntityID(strconv.Itoa(int(rating.RecipeID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = utils.ValidateEntityID(strconv.Itoa(int(rating.UserID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rating.Score < 1 || rating.Score > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "score out of bounds, it must be between 1 and 5"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	err = db.First(&user, rating.UserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user data"})
		return
	}
	err = db.First(&recipe, rating.RecipeID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipe data"})
		return
	}

	err = db.Create(&rating).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add reting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "rating added successfully",
		"id":      rating.ID,
	})
}

func DeleteRatingHandler(c *gin.Context) {
	var rating model.Rating

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

	err = db.First(&rating, validID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "ratin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch rating data"})
		return
	}

	err = db.Delete(&model.Favorite{}, validID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete favorite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rating deleted successfully"})
}
