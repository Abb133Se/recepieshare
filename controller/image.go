package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/service"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
)

// PostUploadRecipeImageHandler godoc
// @Summary Upload image for a recipe
// @Description Uploads an image file and associates it with a recipe by recipe ID
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Recipe ID"
// @Param image formData file true "Image file (jpeg, png, webp)"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid input or file error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipe/{id}/images [post]
func PostUploadRecipeImageHandler(c *gin.Context) {
	recipeID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
// @Description Deletes an image associated with a recipe
// @Tags images
// @Param id path int true "Recipe ID"
// @Param imageId path int true "Image ID"
// @Success 200 {object} map[string]string "Image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipe/{id}/image/{imageId} [delete]
func DeleteRecipeImageHandler(c *gin.Context) {
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

	err = service.DeleteImage("recipe", recipeID, imageID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted successfully"})
}

// PostUploadUserProfileImageHandler godoc
// @Summary Upload user profile image
// @Description Uploads a single profile image for a user
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "User ID"
// @Param image formData file true "Image file (jpeg, png, webp)"
// @Success 200 {object} map[string]interface{} "Image uploaded successfully"
// @Failure 400 {object} map[string]string "Invalid input or file error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/{id}/profile-image [post]
func PostUploadUserProfileImageHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	img, err := service.UploadImage(c, "user", userID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image": img})
}

// GetServeUserProfileImageHandler godoc
// @Summary Serve user profile image by ID
// @Description Retrieves and serves a user's profile image
// @Tags images
// @Produce image/*
// @Param id path int true "User ID"
// @Param imageId path int true "Image ID"
// @Success 200 "Image file served"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 404 {object} map[string]string "Image not found"
// @Router /user/{id}/profile-image/{imageId} [get]
func GetServeUserProfileImageHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageID, err := utils.ValidateEntityID(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.ServeImage(c, "user", userID, imageID, internal.Backend)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
}

// DeleteUserProfileImageHandler godoc
// @Summary Delete user profile image
// @Description Deletes a user's profile image by ID
// @Tags images
// @Param id path int true "User ID"
// @Param imageId path int true "Image ID"
// @Success 200 {object} map[string]string "Profile image deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /user/{id}/profile-image/{imageId} [delete]
func DeleteUserProfileImageHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageID, err := utils.ValidateEntityID(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.DeleteImage("user", userID, imageID, internal.Backend)
	if err != nil {
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
