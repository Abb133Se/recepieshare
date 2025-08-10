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

type TagResponse struct {
	Message string    `json:"message,omitempty"`
	Data    model.Tag `json:"data"`
}

// GetTagHandler godoc
// @Summary      Get a tag by ID
// @Description  Retrieves a tag by its ID
// @Tags         tags
// @Produce      json
// @Param        id   path      int  true  "Tag ID"
// @Success      200  {object}  TagResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tag/{id} [get]
func GetTagHandler(c *gin.Context) {
	var tag model.Tag

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

	err = db.First(&tag, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, TagResponse{Message: "successful", Data: tag})
}

// PostTagHandler godoc
// @Summary      Create a new tag
// @Description  Creates a new tag. Optionally associates it with recipes by IDs.
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        tag  body      model.Tag  true  "Tag data"
// @Success      201  {object}  TagResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tag [post]
func PostTagHandler(c *gin.Context) {
	var tag model.Tag

	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	var existingTag model.Tag
	if err := db.Where("name = ?", tag.Name).First(&existingTag).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "tag already exists"})
		return
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create tag"})
		return
	}

	if len(tag.Recipes) > 0 {
		for _, r := range tag.Recipes {
			var existingRecipe model.Recipe
			if err := db.First(&existingRecipe, r.ID).Error; err == nil {
				db.Model(&tag).Association("Recipes").Append(&existingRecipe)
			}
		}
	}

	c.JSON(http.StatusCreated, TagResponse{Message: "tag created", Data: tag})
}

// GetAllTagsHandler godoc
// @Summary      Get all tags
// @Description  Retrieves all tags with their associated recipes
// @Tags         tags
// @Produce      json
// @Param        limit   query     int     false  "Limit number of recipes returned"
// @Param        offset  query     int     false  "Offset for pagination"
// @Success      200  {array}   TagResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tags [get]
func GetAllTagsHandler(c *gin.Context) {
	var tags []model.Tag

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit := validLimit
	offset := validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	err = db.Preload("Recipes").Limit(limit).Offset(offset).Find(&tags).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}

	c.JSON(http.StatusOK, TagsResponse{Tags: tags})
}

// PutTagHandler godoc
// @Summary      Update a tag by ID
// @Description  Updates the name of a tag specified by its ID
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Tag ID"
// @Param        tag  body      TagsResponse  true  "Updated tag data"
// @Success      200  {object}  TagResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tag/{id} [put]
func PutTagHandler(c *gin.Context) {
	var tag model.Tag
	tagID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag ID"})
		return
	}

	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection failed"})
		return
	}

	var existing model.Tag
	if err := db.First(&existing, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	existing.Name = tag.Name
	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tag"})
		return
	}

	c.JSON(http.StatusOK, TagResponse{Message: "tag updated", Data: existing})
}

// DeleteTagHandler godoc
// @Summary      Delete a tag by ID
// @Description  Deletes a tag and removes all its associations with recipes
// @Tags         tags
// @Produce      json
// @Param        id   path      int  true  "Tag ID"
// @Success      200  {object}  SimpleMessageResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /tag/{id} [delete]
func DeleteTagHandler(c *gin.Context) {
	tagID := c.Param("id")
	var tag model.Tag

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection error"})
		return
	}

	if err := db.Preload("Recipes").First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	if err := db.Model(&tag).Association("Recipes").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tag associations"})
		return
	}

	if err := db.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tag"})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: "tag deleted successfully"})
}
