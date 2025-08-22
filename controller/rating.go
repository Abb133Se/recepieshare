package controller

import (
	"errors"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Request/Response structs for Swagger documentation
type PostRatingRequest struct {
	RecipeID uint `json:"recipe_id" binding:"required"`
	Score    uint `json:"score" binding:"required,min=1,max=5"`
}

type RatingResponse struct {
	Message string `json:"message"`
	ID      uint   `json:"id"`
}

type AverageRatingResponse struct {
	RecipeID uint    `json:"recipe_id"`
	Average  float64 `json:"average"`
	Count    int     `json:"count"`
}

// PostRatingHandler godoc
// @Summary      Add or update a rating for a recipe
// @Description  Adds a new rating for the recipe by the current user (from JWT) or updates the score if one already exists. Score must be between 1 and 5.
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        rating  body      PostRatingRequest  true  "Rating data"
// @Success      200     {object}  RatingResponse
// @Failure      400     {object}  ErrorResponse "Invalid request or score"
// @Failure      401     {object}  ErrorResponse "Unauthorized"
// @Failure      404     {object}  ErrorResponse "Recipe not found"
// @Failure      500     {object}  ErrorResponse "Internal server error"
// @Router       /rating [post]
func PostRatingHandler(c *gin.Context) {
	var req PostRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var recipe model.Recipe
	if err := db.First(&recipe, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
		return
	}

	var existing model.Rating
	if err := db.Where("user_id = ? AND recipe_id = ?", userID, req.RecipeID).
		First(&existing).Error; err == nil {
		existing.Score = req.Score
		if err := db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingUpdateFailed.String()})
			return
		}
		c.JSON(http.StatusOK, RatingResponse{Message: messages.Rating.RatingUpdated.String(), ID: existing.ID})
		return
	}

	rating := model.Rating{UserID: userID, RecipeID: req.RecipeID, Score: req.Score}
	if err := db.Create(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingAddFail.String()})
		return
	}

	c.JSON(http.StatusOK, RatingResponse{Message: messages.Rating.RatingAdded.String(), ID: rating.ID})
}

// DeleteRatingHandler godoc
// @Summary      Delete a rating
// @Description  Deletes a rating by ID if it belongs to the authenticated user (admins can delete any)
// @Tags         ratings
// @Security     BearerAuth
// @Param        id   path      int  true  "Rating ID"
// @Success      200  {object}  SuccessMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /rating/{id} [delete]

// DeleteRatingHandler (admin route) godoc
// @Summary      Delete a rating (admin)
// @Description  Admin deletes a rating of any user
// @Tags         ratings
// @Security     BearerAuth
// @Param        userID path int true "User ID"
// @Param        id     path int true "Rating ID"
// @Success      200 {object} controller.SuccessMessageResponse
// @Failure      400 {object} controller.ErrorResponse
// @Failure      401 {object} controller.ErrorResponse
// @Failure      403 {object} controller.ErrorResponse
// @Failure      404 {object} controller.ErrorResponse
// @Failure      500 {object} controller.ErrorResponse
// @Router       /admin/user/{userID}/rating/{id} [delete]
func DeleteRatingHandler(c *gin.Context) {
	var userID uint
	var err error
	if role := c.GetString("role"); role == "user" {
		userID = c.GetUint("userID")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
			return
		}
	} else if role == "admin" {
		userID, err = middleware.GetEffectiveUserID(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
			return
		}
	}

	ratingID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var rating model.Rating
	if err := db.Where("id = ? AND user_id = ?", ratingID, userID).First(&rating).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: messages.Rating.RatingDeleteForbidden.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	if err := db.Delete(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingDeleteFail.String()})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: messages.Rating.RatingDeleted.String()})
}

// GetAverageRatingHandler godoc
// @Summary      Get average rating for a recipe
// @Description  Retrieve the average rating and count of ratings for a given recipe ID
// @Tags         ratings
// @Param        id   path  int  true  "Recipe ID"
// @Success      200  {object}  AverageRatingResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /recipe/{id}/rating [get]
func GetAverageRatingHandler(c *gin.Context) {
	var ratings []model.Rating

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.Where("recipe_id = ?", validID).Find(&ratings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	if len(ratings) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: messages.Rating.RatingNotFound.String(),
		})
		return
	}

	var total uint
	for _, r := range ratings {
		total += r.Score
	}

	average := float64(total) / float64(len(ratings))

	c.JSON(http.StatusOK, AverageRatingResponse{
		RecipeID: uint(validID),
		Average:  average,
		Count:    len(ratings),
	})
}

// PutUpdateRatingHandler godoc
// @Summary      Update a rating by ID
// @Description  Update the score of an existing rating by its ID
// @Tags         ratings
// @Accept       json
// @Produce      json
// @Param        id      path      int           true  "Rating ID"
// @Param        rating  body      model.Rating  true  "Updated rating data"
// @Success      200     {object}  RatingResponse
// @Failure      400     {object}  ErrorResponse
// @Failure      404     {object}  ErrorResponse
// @Failure      500     {object}  ErrorResponse
// @Router       /rating/{id} [put]
func PutUpdateRatingHandler(c *gin.Context) {
	var rating model.Rating
	var existingRating model.Rating

	validId, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err = c.BindJSON(&rating); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	if rating.Score < 1 || rating.Score > 5 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: messages.Rating.RatingInvalidScore.String()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	err = db.First(&existingRating, validId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Rating.RatingNotFound.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	existingRating.Score = rating.Score
	err = db.Save(&existingRating).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingUpdateFailed.String()})
		return
	}

	c.JSON(http.StatusOK, RatingResponse{
		Message: messages.Rating.RatingUpdated.String(),
		ID:      existingRating.ID,
	})
}

// GetAllRatings godoc
// @Summary      Get all ratings
// @Description  Retrieve a paginated list of all ratings with user and recipe details
// @Tags         ratings
// @Security     BearerAuth
// @Param        limit     query     int     false  "Number of items per page" default(10)
// @Param        offset    query     int     false  "Pagination offset" default(0)
// @Param        sortOrder query     string  false  "Sort order: score_asc, score_desc, date_asc, date_desc" default(date_desc)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /ratings [get]
func GetAllRatings(c *gin.Context) {
	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database connection failed"})
		return
	}

	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}

	// Build query with joins
	query := db.Table("ratings").
		Select(`ratings.id as rating_id,
                ratings.score as score,
                users.id as user_id,
                users.name as user_name,
                recipes.id as recipe_id,
                recipes.title as recipe_title`).
		Joins("JOIN users ON ratings.user_id = users.id").
		Joins("JOIN recipes ON ratings.recipe_id = recipes.id")

	// Count total ratings
	total, err := utils.Count(query, "ratings")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count ratings"})
		return
	}

	// Fetch paginated results
	var ratings []struct {
		RatingID    uint   `json:"rating_id"`
		Score       uint   `json:"score"`
		UserID      uint   `json:"user_id"`
		UserName    string `json:"user_name"`
		RecipeID    uint   `json:"recipe_id"`
		RecipeTitle string `json:"recipe_title"`
	}

	if err := utils.Paginate(query, limit, offset, &ratings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ratings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"ratings": ratings,
	})
}
