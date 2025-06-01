package routes

import (
	"github.com/Abb133Se/recepieshare/controller"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/gin-gonic/gin"
)

func AddRoutes(r *gin.Engine) {
	r.GET("/recipe/:id", controller.GetRecipeHandler)
	r.GET("/recipe/list", controller.GetAllRecipesHandler)
	r.GET("/recipe/:id/ingridients", controller.GetAllRecipeIngridientsHandler)
	r.GET("/recipe/:id/comments", controller.GetAllRecipeCommentsHandler)
	r.GET("/ingridient/:id", controller.GetIngrIdientHandler)
	r.GET("/recipe/:id/rating", controller.GetAverageRatingHandler)
	r.GET("/recipes/top-rated", controller.GetTopRatedRecipesHandler)
	r.GET("/recipes/most-popular", controller.GetMostPopularRecipesHandler)
	r.POST("/forgot-password", controller.ForgotPasswordHandler)
	r.POST("/reset-password", controller.ResetPasswordHandler)
	r.GET("/recipe/:id/calories", controller.GetRecipeNutritionHandler)

	protected := r.Group("/")
	protected.Use(middleware.AuthenticatJWT())
	protected.POST("/recipe", controller.PostRecipeHandler)
	protected.DELETE("/recipe/:id", controller.DeleteRecipeHandler)
	protected.POST("/comment", controller.PostCommentHandler)
	protected.DELETE("/comment/:id", controller.DeleteCommentHandler)
	protected.POST("/ingridient", controller.PostIngridientHandler)
	protected.DELETE("/ingridient/:id", controller.DeleteIngridientHandler)
	protected.GET("/user/:id/recipes", controller.GetUserRecipesHandler)
	protected.GET("/user/:id/favorites", controller.GetUserFavoritesHandler)
	protected.POST("/comment/:id/like", controller.PostCommentLikeIncHandler)
	protected.POST("/comment/:id/dislike", controller.PostCommentLikeDecHandler)
	protected.POST("/favorite", controller.PostFavoriteHandler)
	protected.DELETE("/favorite/:id", controller.DeleteFavoriteHandler)
	protected.POST("/rating", controller.PostRatingHandler)
	protected.DELETE("/rating/:id", controller.DeleteRatingHandler)
	protected.GET("/user/:id/ratings", controller.GetUserRatingsHandler)
	protected.PUT("/recipe/:id", controller.PutRecipeUpdateHandler)
	protected.PUT("/rating/:id", controller.PutUpdateRatingHandler)

	r.POST("/signup", controller.Signup)
	r.POST("/login", controller.Login)
}
