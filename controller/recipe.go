package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/Abb133Se/recepieshare/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TopRatedRecipe struct {
	RecipeID   uint    `json:"recipe_id"`
	Title      string  `json:"title"`
	Average    float64 `joson:"average"`
	TotalVotes int64   `json:"total_votes"`
}

type MostPopularRecipe struct {
	RecipeID      uint   `json:"recipe_id"`
	Title         string `json:"title"`
	FavoriteCount int64  `json:"favorite_count"`
}

var API_KEY = "YTAMecQ6C06ClaR/HmS26g==OUlc0LkiJgLyFjhv"

func GetRecipeHandler(c *gin.Context) {

	var recipe model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to connect to db",
		})
		return
	}

	err = db.Preload("Ingredients").
		Preload("Comments").
		Preload("User").First(&recipe, validID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "recipe not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch recipe",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successful",
		"data":    recipe,
	})
}

func PostRecipeHandler(c *gin.Context) {

	var recipe model.Recipe
	var user model.User

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	if recipe.UserID == 0 && recipe.User.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID or User is required"})
		return
	}

	_, err := utils.ValidateEntityID(strconv.Itoa(int(recipe.UserID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "failed to connect to db",
		})
		return
	}

	err = db.First(&user, recipe.UserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
		return
	}

	result := db.Create(&recipe)
	if result.Error != nil {
		c.JSON(500, gin.H{
			"message": "failed to create record",
		})
		return
	}

	for i := range recipe.Ingredients {
		recipe.Ingredients[i].ID = 0
		recipe.Ingredients[i].RecipeID = recipe.ID
		if err := db.Create(&recipe.Ingredients[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to save ingredient"})
			return
		}
	}

	c.JSON(200, gin.H{
		"message": "recipe created succussfully",
		"id":      recipe.ID,
	})

}

func DeleteRecipeHandler(c *gin.Context) {
	var recepie model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "internal error",
		})
		return
	}

	err = db.First(&recepie, validID).Error
	if err != nil {
		c.JSON(404, gin.H{
			"message": "not found",
		})
		return
	}

	err = db.Delete(&model.Recipe{}, validID).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Failed to delete recipe",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Recipe deleted successfully",
	})
}

func GetAllRecipeIngredientHandler(c *gin.Context) {

	var Ingredient []model.Ingredient

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.First(&model.Recipe{}, validID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "recipe does not exist"})
		return
	}

	err = db.Model(&model.Ingredient{}).Where("recipe_id = ?", c.Param("id")).Find(&Ingredient).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve Ingredient from server"})
		return
	}

	if len(Ingredient) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "recipe does not have Ingredient"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ingredient successfully retrieved",
		"data":    Ingredient,
	})

}

func GetAllRecipeCommentsHandler(c *gin.Context) {
	var comments []model.Comment

	var limit, offset = 1, 0

	id := c.Param("id")

	validID, err := utils.ValidateEntityID(id)
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

	sort := c.DefaultQuery("sort", "")
	var order string
	switch sort {
	case "likes_desc":
		order = "likes DESC"
	case "likes_asc":
		order = "likes ASC"
	default:
		order = ""
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to server"})
		return
	}

	err = db.First(&model.Recipe{}, validID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	query := db.Model(&model.Comment{}).Where("recipe_id = ?", validID).Limit(limit).Offset(offset)
	if order != "" {
		query = query.Order(order)
	}
	err = query.Find(&comments).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comments"})
		return
	}

	var commentCount int64
	err = db.Model(&model.Comment{}).Where("recipe_id = ?", validID).Count(&commentCount).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to count comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{

		"message": "comments retrieved successfully",
		"data":    comments,
		"count":   commentCount,
	})

}

func GetAllRecipesHandler(c *gin.Context) {

	var recipes []model.Recipe
	var limit, offset = 1, 0

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to DB"})
		return
	}

	if err = db.Preload("Comments").Preload("Ingredients").Limit(limit).Offset(offset).Find(&recipes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "recipes retrieved successfully",
		"data":    recipes,
	})

}

func PutRecipeUpdateHandler(c *gin.Context) {
	var input struct {
		Title      string             `json:"title"`
		Text       string             `json:"text"`
		Ingredient []model.Ingredient `json:"ingridient"`
	}

	var recipe model.Recipe

	validID, err := utils.ValidateEntityID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "faild to connect to db"})
		return
	}

	if err = db.First(&recipe, validID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "recipe not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recipe"})
		return
	}

	recipe.Title = input.Title
	recipe.Text = input.Text

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&recipe).Error; err != nil {
			return err
		}

		if err := tx.Where("recipe_id = ?", recipe.ID).Delete(&model.Ingredient{}).Error; err != nil {
			return err
		}

		for _, ing := range input.Ingredient {
			ing.RecipeID = recipe.ID
			if err := tx.Create(&ing).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't update recipe", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "recipe updated successfully"})

}

func GetTopRatedRecipesHandler(c *gin.Context) {
	var results []TopRatedRecipe
	var limit, offset = 1, 0

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

	err = db.Table("ratings").
		Select("recipes.id AS recipe_id, recipes.title, AVG(ratings.score) AS average, COUNT(ratings.id) AS total_votes").
		Joins("JOIN recipes ON recipes.id = ratings.recipe_id").
		Group("ratings.recipe_id").
		Order("average DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch top rated recipes"})
		return
	}

	c.JSON(http.StatusOK, results)

}

func GetMostPopularRecipesHandler(c *gin.Context) {
	var results []MostPopularRecipe
	var limit, offset = 1, 0

	validLimit, validOffset, err := utils.ValidateOffLimit(c.Query("limit"), c.Query("offset"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	limit = validLimit
	offset = validOffset

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to db"})
		return
	}

	err = db.Table("recipes").
		Select("recipes.id as recipe_id, recipes.title, COUNT(favorites.id) as favorite_count").
		Joins("LEFT JOIN favorites ON recipes.id = favorites.recipe_id").
		Group("recipes.id").
		Order("favorite_count DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "faild to fetch popular recipes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"recipes": results})
}

func GetRecipeNutritionHandler(c *gin.Context) {
	id := c.Param("id")

	recipeID, err := utils.ValidateEntityID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid recipe ID"})
		return
	}

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	var ingredients []model.Ingredient
	if err := db.Where("recipe_id = ?", recipeID).Find(&ingredients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch ingredients"})
		return
	}

	var ingredientStrings []string
	for _, ing := range ingredients {
		ingredientStrings = append(ingredientStrings, fmt.Sprintf("%s of %s", ing.Amount, ing.Name))
	}

	nutritionData, err := utils.EstimateNutrition(API_KEY, ingredientStrings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI model inference failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nutritional values": nutritionData})
}
