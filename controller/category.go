package controller

import (
	"errors"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCategoryHandler(c *gin.Context) {
	var category model.Category

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	err = db.First(&category, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "successful", "data": category})
}

func PostCategoryHandler(c *gin.Context) {
	var category model.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	var existingCategory model.Category
	if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "category already exists"})
		return
	}

	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create category"})
		return
	}

	if len(category.Recipes) > 0 {
		for _, r := range category.Recipes {
			var existingRecipe model.Recipe
			if err := db.First(&existingRecipe, r.ID).Error; err == nil {
				db.Model(&category).Association("Recipes").Append(&existingRecipe)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "category created", "data": category})
}

func GetAllCategoriesHandler(c *gin.Context) {
	var categories []model.Category
	var limit, offset = 10, 0

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	sort := c.DefaultQuery("sort", "")
	var order string
	switch sort {
	case "name_desc":
		order = "name DESC"
	case "name_asc":
		order = "name ASC"
	case "created_desc":
		order = "created_at DESC"
	case "created_asc":
		order = "created_at ASC"
	default:
		order = ""
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	query := db.Model(&model.Category{}).Limit(limit).Offset(offset)
	if order != "" {
		query = query.Order(order)
	}

	err = query.Find(&categories).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve categories"})
		return
	}

	var categoryCount int64
	err = db.Model(&model.Category{}).Count(&categoryCount).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to count categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "categories retrieved successfully",
		"data":    categories,
		"count":   categoryCount,
	})
}

func PutCategoryHandler(c *gin.Context) {
	var category model.Category
	categoryID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection failed"})
		return
	}

	var existing model.Category
	if err := db.First(&existing, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	existing.Name = category.Name
	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category updated", "data": existing})
}

func DeleteCategoryHandler(c *gin.Context) {
	categoryID := c.Param("id")
	var category model.Category

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	if err := db.Preload("Recipes").First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	if err := db.Model(&category).Association("Recipes").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove category associations"})
		return
	}

	if err := db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}
