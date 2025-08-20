package controller

import (
	"errors"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagResponse struct {
	Message string    `json:"message,omitempty"`
	Data    model.Tag `json:"data"`
}

type TagsListResponse struct {
	Message string      `json:"message,omitempty"`
	Data    []model.Tag `json:"data"`
	Count   int64       `json:"count"`
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.First(&tag, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Tag.TagNotFound.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}
	c.JSON(http.StatusOK, TagResponse{Message: messages.Common.Success.String(), Data: tag})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var existingTag model.Tag
	if err := db.Where("name = ?", tag.Name).First(&existingTag).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: messages.Tag.TagAlreadyExists.String()})
		return
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Tag.TagCreationFailed.String()})
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

	c.JSON(http.StatusCreated, TagResponse{Message: messages.Tag.TagCreationOk.String(), Data: tag})
}

// GetAllTagsHandler godoc
// @Summary      Get all tags
// @Description  Retrieve all tags, optionally sorted by name or creation date
// @Tags         tags
// @Produce      json
// @Param        sort    query     string  false  "Sort order: name_asc, name_desc, created_asc, created_desc"
// @Success      200     {object}  controller.TagsResponse
// @Failure      500     {object}  controller.ErrorResponse
// @Router       /tags [get]
func GetAllTagsHandler(c *gin.Context) {
	var tags []model.Tag

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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	query := db.Model(&model.Tag{})
	if order != "" {
		query = query.Order(order)
	}

	if err := query.Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Tag.TagFetchFailed.String()})
		return
	}

	c.JSON(http.StatusOK, TagsListResponse{
		Message: messages.Common.Success.String(),
		Data:    tags,
		Count:   int64(len(tags)),
	})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var existing model.Tag
	if err := db.First(&existing, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Tag.TagNotFound.String()})
		return
	}

	existing.Name = tag.Name
	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Tag.TagUploadFailed.String()})
		return
	}

	c.JSON(http.StatusOK, TagResponse{Message: messages.Tag.TagUploadOK.String(), Data: existing})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	if err := db.Preload("Recipes").First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Tag.TagNotFound.String()})
		return
	}

	if err := db.Model(&tag).Association("Recipes").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Tag.TagFailedAssocioationRemova.String()})
		return
	}

	if err := db.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Tag.TagDeletionFaied.String()})
		return
	}

	c.JSON(http.StatusOK, SimpleMessageResponse{Message: messages.Tag.TagDeletionOk.String()})
}
