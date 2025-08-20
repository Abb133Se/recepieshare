package controller

import (
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/messages"
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
	Message string              `json:"message"`
	Data    []FavoriteWithTitle `json:"data"`
	Count   int64               `json:"count"`
}

type FavoriteWithTitle struct {
	model.Favorite
	RecipeTitle string `json:"recipe_title"`
}

type UserRatingsResponse struct {
	Message string         `json:"message"`
	Data    []model.Rating `json:"data"`
	Count   int64          `json:"count"`
}

// UserStats holds computed statistics for a user
// @Description Aggregated statistics based on the user's activity in the system.
type UserStats struct {
	// Total number of recipes created by this user
	TotalRecipes int64 `json:"total_recipes" example:"12"`

	// Total number of comments posted by this user
	TotalComments int64 `json:"total_comments" example:"34"`

	// Total number of times this user's recipes were favorited by others
	TotalFavorites int64 `json:"total_favorites" example:"87"`

	// Average rating across all this user's recipes (scale 1–5)
	AverageRating float64 `json:"average_rating" example:"4.5"`

	// Sum of likes across all comments posted by this user
	TotalLikesOnComments int64 `json:"total_likes_on_comments" example:"56"`

	// Average number of likes per comment posted by this user
	AverageLikesPerComment float64 `json:"average_likes_per_comment" example:"2.3"`

	// Total number of ratings this user has given to other recipes
	TotalRatingsGiven int64 `json:"total_ratings_given" example:"19"`

	// Average rating value this user gives when rating other recipes (scale 1–5)
	AverageRatingGiven float64 `json:"average_rating_given" example:"3.8"`

	// Total number of recipes this user has favorited
	TotalFavoritesGiven int64 `json:"total_favorites_given" example:"25"`
}

// RecipeSummary provides a summary of a recipe used in profile highlights
// @Description Minimal recipe representation used to highlight "most popular" and "highest rated" recipes.
type RecipeSummary struct {
	// Unique identifier of the recipe
	ID uint `json:"id" example:"7"`

	// Title of the recipe
	Title string `json:"title" example:"Classic Pancakes"`

	// Rating score for highest-rated recipe (scale 1–5); only present if relevant
	Score float64 `json:"score,omitempty" example:"4.8"`

	// Favorite count for most-popular recipe; only present if relevant
	Count int64 `json:"count,omitempty" example:"120"`
}

// UserProfileResponse represents the complete profile of a user
// @Description Full user profile including base user data, profile image, aggregated statistics, and highlighted recipes.
type UserProfileResponse struct {
	// Base user information (name, email, etc.)
	// Warning: password and sensitive fields are omitted from this response
	User model.User `json:"user"`

	// Profile image associated with the user (if uploaded), including file path and metadata
	ProfileImage *model.Image `json:"profile_image,omitempty"`

	// Aggregated statistics about user's activity (recipes, comments, favorites, ratings)
	Stats UserStats `json:"stats"`

	// The user's recipe with the highest number of favorites (if exists)
	MostPopular *RecipeSummary `json:"most_popular_recipe,omitempty"`

	// The user's recipe with the highest average rating (if exists)
	HighestRated *RecipeSummary `json:"highest_rated_recipe,omitempty"`
}

// GetUserProfile godoc
// @Summary      Get user profile
// @Description  Fetches the complete profile of the authenticated user.
// @Description  The response includes:
// @Description  - **Base user info** (ID, name, email, timestamps)
// @Description  - **Profile image** (optional, if uploaded)
// @Description  - **User statistics** (totals, averages, likes, ratings, favorites)
// @Description  - **Highlights** (most popular and highest-rated recipes created by the user)
// @Tags         users
// @Produce      json
// @Success      200 {object} UserProfileResponse "User profile with details, stats, and highlights"
// @Failure      401 {object} ErrorResponse "Unauthorized: missing or invalid token"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /user/profile [get]
func GetUserProfile(c *gin.Context) {
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

	// --- 1. Load user + relationships ---
	var user model.User
	if err := db.Preload("Recipes.Ingredients").
		Preload("Recipes.Comments").
		Preload("Recipes.Favorites").
		Preload("Recipes.Ratings").
		Preload("Recipes.Tags").
		Preload("Recipes.Categories").
		Preload("Recipes.Steps").
		Preload("Comments").
		Preload("Favorites").
		Preload("Ratings").
		First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// sanitize sensitive fields
	user.Password = ""
	user.Salt = ""
	user.PasswordResetToken = ""
	user.PasswordResetExpiresAt = nil

	// --- 2. Fetch profile image ---
	var profileImage model.Image
	if err := db.Where("entity_type = ? AND entity_id = ?", "user", userID).
		First(&profileImage).Error; err != nil {
		profileImage = model.Image{}
	}

	// --- 3. Aggregated statistics in one query ---
	var stats UserStats
	err = db.Raw(`
		SELECT
			(SELECT COUNT(*) FROM recipes r WHERE r.user_id = ?) AS total_recipes,
			(SELECT COUNT(*) FROM comments c WHERE c.user_id = ?) AS total_comments,
			(SELECT COUNT(*) FROM favorites f JOIN recipes r2 ON f.recipe_id = r2.id WHERE r2.user_id = ?) AS total_favorites,
			(SELECT COALESCE(AVG(rt.score),0) FROM ratings rt JOIN recipes r3 ON rt.recipe_id = r3.id WHERE r3.user_id = ?) AS average_rating,
			(SELECT COALESCE(SUM(c2.likes),0) FROM comments c2 WHERE c2.user_id = ?) AS total_likes_on_comments,
			(SELECT COALESCE(AVG(c3.likes),0) FROM comments c3 WHERE c3.user_id = ?) AS average_likes_per_comment,
			(SELECT COUNT(*) FROM ratings r4 WHERE r4.user_id = ?) AS total_ratings_given,
			(SELECT COALESCE(AVG(r5.score),0) FROM ratings r5 WHERE r5.user_id = ?) AS average_rating_given,
			(SELECT COUNT(*) FROM favorites f2 WHERE f2.user_id = ?) AS total_favorites_given
	`, userID, userID, userID, userID, userID, userID, userID, userID, userID).Scan(&stats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.User.UserStatFetchFail.String()})
		return
	}

	// --- 4. Most popular recipe (highest favorites count) ---
	var mostPopular RecipeSummary
	db.Raw(`
		SELECT r.id, r.title, COUNT(f.id) AS count
		FROM recipes r
		LEFT JOIN favorites f ON f.recipe_id = r.id
		WHERE r.user_id = ?
		GROUP BY r.id
		ORDER BY COUNT(f.id) DESC
		LIMIT 1
	`, userID).Scan(&mostPopular)

	// --- 5. Highest-rated recipe (best average rating) ---
	var highestRated RecipeSummary
	db.Raw(`
		SELECT r.id, r.title, COALESCE(AVG(rt.score),0) AS score
		FROM recipes r
		LEFT JOIN ratings rt ON rt.recipe_id = r.id
		WHERE r.user_id = ?
		GROUP BY r.id
		ORDER BY AVG(rt.score) DESC
		LIMIT 1
	`, userID).Scan(&highestRated)

	// --- 6. Build response ---
	resp := UserProfileResponse{
		User:         user,
		Stats:        stats,
		MostPopular:  nil,
		HighestRated: nil,
	}

	if profileImage.ID != 0 {
		resp.ProfileImage = &profileImage
	}
	if mostPopular.ID != 0 {
		resp.MostPopular = &mostPopular
	}
	if highestRated.ID != 0 {
		resp.HighestRated = &highestRated
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserRecipesHandler godoc
// @Summary      Get user's recipes with pagination
// @Description  Retrieve a paginated list of recipes for the logged-in user with total count
// @Tags         users
// @Produce      json
// @Param        limit  query     int     false  "Limit number of recipes returned"
// @Param        offset query     int     false  "Number of recipes to skip"
// @Success      200    {object}  controller.UserRecipesResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/recipes [get]
func GetUserRecipesHandler(c *gin.Context) {
	// ✅ Get userID from JWT middleware
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
		return
	}
	userID := userIDValue.(uint)

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	query := db.Model(&model.Recipe{}).Where("user_id = ?", userID)

	var recipes []model.Recipe
	totalCount, err := utils.PaginateAndCount(c, query, &recipes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Recipe.RecipeFetchFail.String()})
		return
	}

	c.JSON(http.StatusOK, UserRecipesResponse{
		Message: messages.Common.Success.String(),
		Data:    recipes,
		Count:   totalCount,
	})
}

// GetUserFavoritesHandler godoc
// @Summary      Get user's favorites with pagination
// @Description  Retrieve a paginated list of favorite recipes for the logged-in user
// @Tags         users
// @Produce      json
// @Param        limit  query     int     false  "Limit number of favorites returned"
// @Param        offset query     int     false  "Number of favorites to skip"
// @Success      200    {object}  controller.UserFavoritesResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/favorites [get]
func GetUserFavoritesHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
		return
	}
	userID := userIDValue.(uint)

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	query := db.Table("favorites").
		Select("favorites.id, favorites.recipe_id, recipes.title as recipe_title").
		Joins("JOIN recipes ON favorites.recipe_id = recipes.id").
		Where("favorites.user_id = ?", userID)

	var favorites []FavoriteWithTitle
	totalCount, err := utils.PaginateAndCount(c, query, &favorites)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Favorite.FavoriteFailed.String()})
		return
	}

	c.JSON(http.StatusOK, UserFavoritesResponse{
		Message: messages.Common.Success.String(),
		Data:    favorites,
		Count:   totalCount,
	})
}

// GetUserRatingsHandler godoc
// @Summary      Get user's ratings with pagination
// @Description  Retrieve a paginated list of ratings for the logged-in user
// @Tags         users
// @Produce      json
// @Param        limit  query     int     false  "Limit number of ratings returned"
// @Param        offset query     int     false  "Number of ratings to skip"
// @Success      200    {object}  controller.UserRatingsResponse
// @Failure      400    {object}  controller.ErrorResponse
// @Failure      500    {object}  controller.ErrorResponse
// @Router       /user/ratings [get]
func GetUserRatingsHandler(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: messages.Common.Unauthorized.String()})
		return
	}
	userID := userIDValue.(uint)

	db, err := internal.GetGormInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Common.DBConnectionErr.String()})
		return
	}

	query := db.Model(&model.Rating{}).Where("user_id = ?", userID)

	var ratings []model.Rating
	totalCount, err := utils.PaginateAndCount(c, query, &ratings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: messages.Rating.RatingFetchFail.String()})
		return
	}

	c.JSON(http.StatusOK, UserRatingsResponse{
		Message: messages.Common.Success.String(),
		Data:    ratings,
		Count:   totalCount,
	})
}
