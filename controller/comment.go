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

type CommentResponse struct {
	Message string `json:"message"`
	ID      uint   `json:"id"`
}

type PostCommentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	RecipeID    uint   `json:"recipe_id" binding:"required"`
}

// PostCommentHandler godoc
// @Summary      Post a new comment on a recipe
// @Description  Creates a new comment linked to a recipe by the current user (from JWT). Each user can comment only once per recipe.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        comment  body      PostCommentRequest  true  "Comment data"
// @Success      200      {object}  CommentResponse
// @Failure      400      {object}  ErrorResponse "Invalid request"
// @Failure      401      {object}  ErrorResponse "Unauthorized"
// @Failure      404      {object}  ErrorResponse "Recipe not found"
// @Failure      409      {object}  ErrorResponse "User has already commented on this recipe"
// @Failure      500      {object}  ErrorResponse "Internal server error"
// @Router       /comment [post]
func PostCommentHandler(c *gin.Context) {
	var req PostCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	var recipe model.Recipe
	if err := db.First(&recipe, req.RecipeID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "recipe not found"})
		return
	}

	var existingComment model.Comment
	if err := db.Where("user_id = ? AND recipe_id = ?", userID, req.RecipeID).
		First(&existingComment).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "you have already commented on this recipe"})
		return
	}

	comment := model.Comment{
		Title:       req.Title,
		Description: req.Description,
		RecipeID:    req.RecipeID,
		UserID:      userID,
	}
	if err := db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create comment"})
		return
	}

	c.JSON(http.StatusOK, CommentResponse{
		Message: "comment successfully posted",
		ID:      comment.ID,
	})
}

// DeleteCommentHandler godoc
// @Summary      Delete a comment by ID
// @Description  Delete a comment if it belongs to the authenticated user
// @Tags         comments
// @Security     BearerAuth
// @Param        id   path  int  true  "Comment ID"
// @Success      200  {object}  SuccessMessageResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /comment/{id} [delete]
func DeleteCommentHandler(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}

	commentID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	var comment model.Comment
	if err := db.Where("id = ? AND user_id = ?", commentID, userID).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "not allowed to delete this comment"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	if err := db.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: "comment deleted successfully"})
}

// PostCommentLikeIncHandler godoc
// @Summary      Increment comment like count
// @Description  Increase likes for a comment by ID
// @Tags         comments
// @Param        id   path  int  true  "Comment ID"
// @Success      200  {object}  controller.SuccessMessageResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /comment/{id}/like/inc [post]
func PostCommentLikeIncHandler(c *gin.Context) {
	commentID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	result := db.Exec("update comments set likes = likes + 1 where id = ?", commentID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to like comment"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "comment not found"})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: "comment liked successfully"})
}

// PostCommentLikeDecHandler godoc
// @Summary      Decrement comment like count
// @Description  Decrease likes for a comment by ID
// @Tags         comments
// @Param        id   path  int  true  "Comment ID"
// @Success      200  {object}  controller.SuccessMessageResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /comment/{id}/like/dec [post]
func PostCommentLikeDecHandler(c *gin.Context) {
	commentID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	result := db.Exec("update comments set likes = likes - 1 where id = ?", commentID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to dislike comment"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "comment not found"})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: "comment disliked successfully"})
}
