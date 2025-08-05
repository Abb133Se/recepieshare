package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetIngredientHandler(c *gin.Context) {

	var Ingredient model.Ingredient

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal error",
		})
		return
	}

	err = db.First(&Ingredient, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "internal server error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
		"data":    Ingredient,
	})
}

func PostIngredientHandler(c *gin.Context) {

	var Ingredient model.Ingredient
	var recipe model.Recipe

	err := c.BindJSON(&Ingredient)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
		return
	}

	if Ingredient.ID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(Ingredient.ID))); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if Ingredient.RecipeID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(Ingredient.RecipeID))); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to connect to db",
		})
		return
	}

	err = db.First(&recipe, Ingredient.RecipeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = db.Create(&Ingredient).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to creat Ingredient"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "recorde successfully inserted",
		"id":      Ingredient.ID,
	})

}

func DeleteIngredientHandler(c *gin.Context) {

	var Ingredient model.Ingredient

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "intenal server error"})
		return
	}

	err = db.First(&Ingredient, validID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	err = db.Delete(&model.Ingredient{}, validID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "record deleted successfullly"})
}
