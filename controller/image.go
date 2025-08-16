package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/service"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
)

// PostUploadRecipeImageHandler godoc
// @Summary Upload image for a recipe
// @Description Uploads an image file and associates it with a recipe (must belong to the authenticated user)
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Recipe ID"
// @Param image formData file true "Image file (jpeg, png, webp)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid input or file error"
// @Failure 403 {object} map[string]string "Not authorized to upload for this recipe"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipe/{id}/image [post]
func PostUploadRecipeImageHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := internal.GetGormInstance()
	var recipe model.Recipe
	if err := db.First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	if recipe.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to upload image for this recipe"})
		return
	}

	img, err := service.UploadImage(c, "recipe", recipeID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"image": img})
}

// GetServeRecipeImageHandler godoc
// @Summary Serve recipe image by ID
// @Description Retrieves and serves an image associated with a recipe
// @Tags images
// @Produce image/*
// @Param id path int true "Recipe ID"
// @Param imageId path int true "Image ID"
// @Success 200 "Image file served"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 404 {object} map[string]string "Image not found"
// @Router /recipe/{id}/image/{imageId} [get]
func GetServeRecipeImageHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageID, err := utils.ValidateEntityID(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.ServeImage(c, "recipe", recipeID, imageID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
}

// DeleteRecipeImageHandler godoc
// @Summary Delete a recipe image by ID
// @Description Deletes an image associated with a recipe (must belong to the authenticated user)
// @Tags images
// @Param id path int true "Recipe ID"
// @Param imageId path int true "Image ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 403 {object} map[string]string "Not authorized to delete this recipe image"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipe/{id}/image/{imageId} [delete]
func DeleteRecipeImageHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	recipeID, _ := utils.ValidateEntityID(c.Param("id"))
	imageID, _ := utils.ValidateEntityID(c.Param("imageId"))

	db, _ := internal.GetGormInstance()
	var recipe model.Recipe
	if err := db.First(&recipe, recipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
		return
	}
	if recipe.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not authorized to delete image for this recipe"})
		return
	}

	if err := service.DeleteImage("recipe", recipeID, imageID, internal.Backend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "image deleted successfully"})
}

// PostUploadUserProfileImageHandler godoc
// @Summary Upload user profile image
// @Description Uploads a single profile image for the authenticated user
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file (jpeg, png, webp)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid input or file error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/profile-image [post]
func PostUploadUserProfileImageHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	img, err := service.UploadImage(c, "user", userID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"image": img})
}

// GetServeUserProfileImageHandler godoc
// @Summary Serve authenticated user's profile image
// @Description Retrieves and serves a user's profile image
// @Tags images
// @Produce image/*
// @Param imageId path int true "Image ID"
// @Security BearerAuth
// @Success 200 "Image file served"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 404 {object} map[string]string "Image not found"
// @Router /user/profile-image/{imageId} [get]
func GetServeUserProfileImageHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	imageID, _ := utils.ValidateEntityID(c.Param("imageId"))

	if err := service.ServeImage(c, "user", userID, imageID, internal.Backend); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
}

// DeleteUserProfileImageHandler godoc
// @Summary Delete authenticated user's profile image
// @Description Deletes a user's profile image by ID
// @Tags images
// @Param imageId path int true "Image ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Profile image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/profile-image/{imageId} [delete]
func DeleteUserProfileImageHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	imageID, _ := utils.ValidateEntityID(c.Param("imageId"))

	if err := service.DeleteImage("user", userID, imageID, internal.Backend); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile image deleted successfully"})
}

// GetImageHandler godoc
// @Summary Serve image by entity type and ID
// @Description Generic endpoint to serve images for any entity type
// @Tags images
// @Produce image/*
// @Param entity path string true "Entity type (e.g. user, recipe)"
// @Param entityId path int true "Entity ID"
// @Param imageId path int true "Image ID"
// @Success 200 "Image file served"
// @Failure 400 {object} map[string]string "Invalid ID or entity parameters"
// @Failure 404 {object} map[string]string "Image not found"
// @Router /image/{entity}/{entityId}/{imageId} [get]
func GetImageHandler(c *gin.Context) {
	entityType := c.Param("entity")

	entityID, err := utils.ValidateEntityID(c.Param("entityId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageID, err := utils.ValidateEntityID(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.ServeImage(c, entityType, entityID, imageID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
}
