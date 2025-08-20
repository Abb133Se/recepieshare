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

// Response structs for Swagger

type CategoryResponse struct {
	Message string         `json:"message"`
	Data    model.Category `json:"data"`
}

type CategoriesResponse struct {
	Message string           `json:"message"`
	Data    []model.Category `json:"data"`
	Count   int64            `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessMessageResponse struct {
	Message string `json:"message"`
}

// GetCategoryHandler godoc
// @Summary      Get a category by ID
// @Description  Retrieve a single category by its ID
// @Tags         categories
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  controller.CategoryResponse
// @Failure      400  {object}  controller.ErrorResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /category/{id} [get]
func GetCategoryHandler(c *gin.Context) {
	var category model.Category

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

	err = db.First(&category, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Recipe.RecipeNotFound.String()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}
	c.JSON(http.StatusOK, CategoryResponse{Message: messages.Common.Success.String(), Data: category})
}

// PostCategoryHandler godoc
// @Summary      Create a new category
// @Description  Create a category with optional associated recipes by IDs
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category  body      model.Category  true  "Category data"
// @Success      201       {object}  controller.CategoryResponse
// @Failure      400       {object}  controller.ErrorResponse
// @Failure      409       {object}  controller.ErrorResponse
// @Failure      500       {object}  controller.ErrorResponse
// @Router       /category [post]
func PostCategoryHandler(c *gin.Context) {
	var category model.Category

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "bad request"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var existingCategory model.Category
	if err := db.Where("name = ?", category.Name).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict, ErrorResponse{Error: messages.Category.CatAlreadyExists.String()})
		return
	}

	if err := db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatCreationFailed.String()})
		return
	}

	if len(category.Recipes) > 0 {
		for _, r := range category.Recipes {
			var existingRecipe model.Recipe
			if err := db.First(&existingRecipe, r.ID).Error; err == nil {
				db.Model(&category).Association("Recipes").Append(&existingRecipe)
			}
		}
	}

	c.JSON(http.StatusCreated, CategoryResponse{Message: messages.Category.CatCreationOk.String(), Data: category})
}

// GetAllCategoriesHandler godoc
// @Summary      Get all categories with pagination and sorting
// @Description  Retrieve a paginated list of categories with total count, optionally sorted by name or creation date
// @Tags         categories
// @Produce      json
// @Param        limit   query     int     false  "Limit number of categories returned" default(10)
// @Param        offset  query     int     false  "Number of categories to skip" default(0)
// @Param        sort    query     string  false  "Sort order: name_asc, name_desc, created_asc, created_desc"
// @Success      200     {object}  controller.CategoriesResponse
// @Failure      400     {object}  controller.ErrorResponse
// @Failure      500     {object}  controller.ErrorResponse
// @Router       /categories [get]
func GetAllCategoriesHandler(c *gin.Context) {
	var categories []model.Category

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

	query := db.Model(&model.Category{})
	if order != "" {
		query = query.Order(order)
	}

	// Use modular func: Validates, counts, paginates, fetches.
	queryCount, err := utils.Count(query, "categories")
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Comment.CommnetFetchFail.String()})
		return
	}

	limit, offset, _ := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if limit == 0 {
		limit = 10
	}

	if err := utils.Paginate(query, limit, offset, &categories); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatFetchFailed.String()})
		return
	}

	// totalCount, err := utils.PaginateAndCount(c, query, &categories)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatFetchFailed.String()})
	// 	return
	// }

	c.JSON(http.StatusOK, CategoriesResponse{
		Message: messages.Common.Success.String(),
		Data:    categories,
		Count:   queryCount, // Now guaranteed.
	})
}

// PutCategoryHandler godoc
// @Summary      Update a category by ID
// @Description  Update category details by its ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id        path      int             true  "Category ID"
// @Param        category  body      model.Category  true  "Updated category data"
// @Success      200       {object}  controller.CategoryResponse
// @Failure      400       {object}  controller.ErrorResponse
// @Failure      404       {object}  controller.ErrorResponse
// @Failure      500       {object}  controller.ErrorResponse
// @Router       /category/{id} [put]
func PutCategoryHandler(c *gin.Context) {
	var category model.Category
	categoryID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid category ID"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	var existing model.Category
	if err := db.First(&existing, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Category.CatNotFound.String()})
		return
	}

	existing.Name = category.Name
	if err := db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatUpdateFail.String()})
		return
	}

	c.JSON(http.StatusOK, CategoryResponse{Message: messages.Category.CatUpdateOk.String(), Data: existing})
}

// DeleteCategoryHandler godoc
// @Summary      Delete a category by ID
// @Description  Delete a category and clear its recipe associations
// @Tags         categories
// @Param        id   path  int  true  "Category ID"
// @Success      200  {object}  controller.SuccessMessageResponse
// @Failure      404  {object}  controller.ErrorResponse
// @Failure      500  {object}  controller.ErrorResponse
// @Router       /category/{id} [delete]
func DeleteCategoryHandler(c *gin.Context) {
	categoryID := c.Param("id")
	var category model.Category

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	if err := db.Preload("Recipes").First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: messages.Category.CatNotFound.String()})
		return
	}

	if err := db.Model(&category).Association("Recipes").Clear(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatFailedAssocioationRemova.String()})
		return
	}

	if err := db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Category.CatDeletionFaied.String()})
		return
	}

	c.JSON(http.StatusOK, SuccessMessageResponse{Message: messages.Category.CatDeletionOk.String()})
}
