package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
)

type UserRecipesResponse struct {
	Message string         `json:"message"`
	Data    []model.Recipe `json:"data"`
	Count   int64          `json:"count"`
}

type UserFavoritesResponse struct {
	Message string           `json:"message"`
	Data    []model.Favorite `json:"data"`
	Count   int64            `json:"count"`
}

type UserRatingsResponse struct {
	Message string         `json:"message"`
	Data    []model.Rating `json:"data"`
	Count   int64          `json:"count"`
}

// GetUserRecipesHandler godoc
// @Summary      Get user's recipes with pagination
// @Description  Retrieve a paginated list of recipes for a specific user with total count
// @Tags         users
// @Produce      json
// @Param        id     path      int     true   "User ID"
// @Param        limit  query     int     false  "Limit number of recipes returned"
// @Param        offset query     int     false  "Number of recipes to skip"
// @Success      200    {object}  controller.UserRecipesResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/{id}/recipes [get]
func GetUserRecipesHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	query := db.Model(&model.Recipe{}).Where("user_id = ?", userID)

	var recipes []model.Recipe
	totalCount, err := utils.PaginateAndCount(c, query, &recipes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve recipes"})
		return
	}

	c.JSON(http.StatusOK, UserRecipesResponse{
		Message: "recipes retrieved successfully",
		Data:    recipes,
		Count:   totalCount,
	})
}

// GetUserFavoritesHandler godoc
// @Summary      Get user's favorites with pagination
// @Description  Retrieve a paginated list of favorite recipes for a specific user with total count
// @Tags         users
// @Produce      json
// @Param        id     path      int     true   "User ID"
// @Param        limit  query     int     false  "Limit number of favorites returned"
// @Param        offset query     int     false  "Number of favorites to skip"
// @Success      200    {object}  controller.UserFavoritesResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/{id}/favorites [get]
func GetUserFavoritesHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	query := db.Model(&model.Favorite{}).Where("user_id = ?", userID)

	var favorites []model.Favorite
	totalCount, err := utils.PaginateAndCount(c, query, &favorites)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve favorites"})
		return
	}

	c.JSON(http.StatusOK, UserFavoritesResponse{
		Message: "favorites retrieved successfully",
		Data:    favorites,
		Count:   totalCount,
	})
}

// GetUserRatingsHandler godoc
// @Summary      Get user's ratings with pagination
// @Description  Retrieve a paginated list of ratings by a specific user with total count
// @Tags         users
// @Produce      json
// @Param        id     path      int     true   "User ID"
// @Param        limit  query     int     false  "Limit number of ratings returned"
// @Param        offset query     int     false  "Number of ratings to skip"
// @Success      200    {object}  controller.UserRatingsResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/{id}/ratings [get]
func GetUserRatingsHandler(c *gin.Context) {
	userID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	query := db.Model(&model.Rating{}).Where("user_id = ?", userID)

	var ratings []model.Rating
	totalCount, err := utils.PaginateAndCount(c, query, &ratings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve ratings"})
		return
	}

	c.JSON(http.StatusOK, UserRatingsResponse{
		Message: "ratings retrieved successfully",
		Data:    ratings,
		Count:   totalCount,
	})
}
