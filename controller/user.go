package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserRecipesResponse struct {
	Message string         `json:"message"`
	Data    []model.Recipe `json:"data"`
}

type UserFavoritesResponse struct {
	Message string           `json:"message"`
	Data    []model.Favorite `json:"data"`
}

type UserRatingsResponse struct {
	Message string         `json:"message"`
	Data    []model.Rating `json:"data"`
}

// GetUserRecipesHandler godoc
// @Summary      Get recipes created by a user
// @Description  Retrieves recipes authored by the specified user, supports pagination via limit and offset query params
// @Tags         users
// @Produce      json
// @Param        id      path      int     true   "User ID"
// @Param        limit   query     int     false  "Limit number of recipes returned"
// @Param        offset  query     int     false  "Offset for pagination"
// @Success      200     {object}  UserRecipesResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /user/{id}/recipes [get]
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

	err = db.Preload("Ingredients").Preload("Comments").
		Where("user_id = ?", validID).
		Limit(limit).
		Offset(offset).
		Find(&recipes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipes"})
		return
	}

	c.JSON(http.StatusOK, UserRecipesResponse{
		Message: "recipes fetcjed successfully",
		Data:    recipes,
	})

}

// GetUserFavoritesHandler godoc
// @Summary      Get favorite recipes of a user
// @Description  Retrieves favorite recipes of the specified user, supports pagination via limit and offset
// @Tags         users
// @Produce      json
// @Param        id      path      int     true   "User ID"
// @Param        limit   query     int     false  "Limit number of favorites returned"
// @Param        offset  query     int     false  "Offset for pagination"
// @Success      200     {object}  UserFavoritesResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /user/{id}/favorites [get]
func GetUserFavoritesHandler(c *gin.Context) {
	var favorites []model.Favorite

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

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	if err = db.First(&model.User{}, validID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user from db"})
		return
	}

	err = db.Where("user_id = ?", validID).
		Limit(validLimit).
		Offset(validOffset).
		Find(&favorites).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch favorites"})
		return
	}

	c.JSON(http.StatusOK, UserFavoritesResponse{
		Message: "favorites fetched successfully",
		Data:    favorites,
	})

}

// GetUserRatingsHandler godoc
// @Summary      Get ratings given by a user
// @Description  Retrieves ratings provided by the specified user, supports pagination via limit and offset
// @Tags         users
// @Produce      json
// @Param        id      path      int     true   "User ID"
// @Param        limit   query     int     false  "Limit number of ratings returned"
// @Param        offset  query     int     false  "Offset for pagination"
// @Success      200     {object}  UserRatingsResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /user/{id}/ratings [get]
func GetUserRatingsHandler(c *gin.Context) {
	var ratings []model.Rating

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

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	err = db.First(&model.User{}, validID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"erro": "user not found"})

			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user data"})
		return
	}

	err = db.Where("user_id = ?", validID).
		Offset(validOffset).
		Limit(validLimit).
		Find(&ratings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ratings"})
		return
	}

	c.JSON(http.StatusOK, UserRatingsResponse{
		Message: "ratings fetched successfully",
		Data:    ratings,
	})
}
