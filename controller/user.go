package controller

import (
	"fmt"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserRecipesHandler(c *gin.Context) {
	var recipes []model.Recipe
	var limit, offset = 1, 0

	validID, err := utils.ValidateEntityID(c.Param("id"))
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

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.First(&model.User{}, validID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	fmt.Println("======================")
	fmt.Println(recipes)

	err = db.Preload("Ingridients").Preload("Comments").
		Where("user_id = ?", validID).
		Limit(limit).
		Offset(offset).
		Find(&recipes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipes"})
		return
	}

	fmt.Println("======================")
	fmt.Println(recipes)
	c.JSON(http.StatusOK, gin.H{
		"message": "recipes fetcjed successfully",
		"data":    recipes,
	})

}
