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

// Response structs for Swagger documentation

type IngredientResponse struct {
	Message string           `json:"message"`
	Data    model.Ingredient `json:"data"`
}

type IngredientCreateResponse struct {
	Message string `json:"message"`
	ID      uint   `json:"id"`
}

// GetIngredientHandler godoc
// @Summary      Get an ingredient by ID
// @Description  Retrieve an ingredient by its ID
// @Tags         ingredients
// @Param        id   path      int  true  "Ingredient ID"
// @Success      200  {object}  controller.IngredientResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /ingredient/{id} [get]
func GetIngredientHandler(c *gin.Context) {
	var ingredient model.Ingredient

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal error"})
		return
	}

	err = db.First(&ingredient, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}
	c.JSON(http.StatusOK, IngredientResponse{
		Message: "successful",
		Data:    ingredient,
	})
}

// PostIngredientHandler godoc
// @Summary      Create a new ingredient
// @Description  Create an ingredient linked to a recipe
// @Tags         ingredients
// @Accept       json
// @Produce      json
// @Param        ingredient  body      model.Ingredient  true  "Ingredient data"
// @Success      201         {object}  controller.IngredientCreateResponse
// @Failure      400         {object}  controller.ErrorResponse
// @Failure      404         {object}  controller.ErrorResponse
// @Failure      500         {object}  controller.ErrorResponse
// @Router       /ingredient [post]
func PostIngredientHandler(c *gin.Context) {
	var ingredient model.Ingredient
	var recipe model.Recipe

	err := c.BindJSON(&ingredient)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	if ingredient.ID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(ingredient.ID))); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if ingredient.RecipeID != 0 {
		if _, err := utils.ValidateEntityID(strconv.Itoa(int(ingredient.RecipeID))); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to connect to db"})
		return
	}

	err = db.First(&recipe, ingredient.RecipeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	err = db.Create(&ingredient).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create ingredient"})
		return
	}

	c.JSON(http.StatusCreated, IngredientCreateResponse{
		Message: "record successfully inserted",
		ID:      ingredient.ID,
	})
}

// DeleteIngredientHandler godoc
// @Summary      Delete an ingredient by ID
// @Description  Delete an ingredient record by its ID
// @Tags         ingredients
// @Param        id   path  int  true  "Ingredient ID"
// @Success      200  {object}  controller.SuccessMessageResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /ingredient/{id} [delete]
func DeleteIngredientHandler(c *gin.Context) {
	var ingredient model.Ingredient

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	err = db.First(&ingredient, validID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}

	err = db.Delete(&model.Ingredient{}, validID).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: "record deleted successfully"})
}
