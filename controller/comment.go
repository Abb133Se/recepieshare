package controller

import (
	"errors"
	"net/http"
	"strconv"

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

// PostCommentHandler godoc
// @Summary      Post a new comment
// @Description  Create a new comment linked to a recipe and user
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        comment  body      model.Comment  true  "Comment data"
// @Success      200      {object}  controller.CommentResponse
// @Failure      400      {object}  controller.ErrorResponse
// @Failure      404      {object}  controller.ErrorResponse
// @Failure      500      {object}  controller.ErrorResponse
// @Router       /comment [post]
func PostCommentHandler(c *gin.Context) {
	var comment model.Comment
	var recipe model.Recipe
	var user model.User

	err := c.BindJSON(&comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	if comment.ID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(comment.ID))); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if comment.RecipeID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(comment.RecipeID))); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if comment.UserID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(comment.UserID))); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	err = db.First(&recipe, comment.RecipeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve recipe from server"})
		return
	}

	err = db.First(&user, comment.UserID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve user from server"})
		return
	}

	err = db.Create(&comment).Error
	if err != nil {
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
// @Description  Delete a comment by its ID
// @Tags         comments
// @Param        id   path  int  true  "Comment ID"
// @Success      200  {object}  controller.SuccessMessageResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /comment/{id} [delete]
func DeleteCommentHandler(c *gin.Context) {
	var comment model.Comment

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to server"})
		return
	}

	err = db.First(&comment, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	err = db.Delete(&model.Comment{}, validID).Error
	if err != nil {
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
