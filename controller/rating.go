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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	var recipe model.Recipe
	if err := db.First(&recipe, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "recipe not found"})
		return
	}

	var existing model.Rating
	if err := db.Where("user_id = ? AND recipe_id = ?", userID, req.RecipeID).
		First(&existing).Error; err == nil {
		existing.Score = req.Score
		if err := db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update rating"})
			return
		}
		c.JSON(http.StatusOK, RatingResponse{Message: "rating updated successfully", ID: existing.ID})
		return
	}

	rating := model.Rating{UserID: userID, RecipeID: req.RecipeID, Score: req.Score}
	if err := db.Create(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to add rating"})
		return
	}

	c.JSON(http.StatusOK, RatingResponse{Message: "rating added successfully", ID: rating.ID})
}

// DeleteRatingHandler godoc
// @Summary      Delete a rating by ID
// @Description  Delete a rating if it belongs to the authenticated user
// @Tags         ratings
// @Security     BearerAuth
// @Param        id   path  int  true  "Rating ID"
// @Success      200  {object}  SuccessMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /rating/{id} [delete]
func DeleteRatingHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	ratingID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	var rating model.Rating
	if err := db.Where("id = ? AND user_id = ?", ratingID, userID).First(&rating).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "not allowed to delete this rating"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch rating data"})
		return
	}

	if err := db.Delete(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete rating"})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: "rating deleted successfully"})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	err = db.Where("recipe_id = ?", validID).Find(&ratings).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch ratings"})
		return
	}

	if len(ratings) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "no ratings found for the given recipe",
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "rating score must be between 1 and 5"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	err = db.First(&existingRating, validId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "rating not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch rating"})
		return
	}

	existingRating.Score = rating.Score
	err = db.Save(&existingRating).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update rating"})
		return
	}

	c.JSON(http.StatusOK, RatingResponse{
		Message: "rating updated successfully",
		ID:      existingRating.ID,
	})
}
