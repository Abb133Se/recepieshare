package routes

import (
	"github.com/Abb133Se/recepieshare/controller"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func AddRoutes(r *gin.Engine) {
	public := r.Group("/")
	{
		// Recipe read endpoints
		public.GET("/recipe/:id", middleware.ExtractUserFromToken(), controller.GetRecipeHandler)
		public.GET("/recipe/list", controller.GetAllRecipesHandler)
		public.GET("/recipe/:id/ingredients", controller.GetAllRecipeIngredientHandler)
		public.GET("/recipe/:id/comments", controller.GetAllRecipeCommentsHandler)
		public.GET("/recipe/:id/rating", controller.GetAverageRatingHandler)
		public.GET("/recipe/:id/calories", controller.GetRecipeNutritionHandler)
		public.GET("/recipes/top-rated", controller.GetTopRatedRecipesHandler)
		public.GET("/recipes/most-popular", controller.GetMostPopularRecipesHandler)
		public.GET("/recipes/search", controller.SearchRecipesHandler)

		// Ingredient read endpoint
		public.GET("/ingredient/:id", controller.GetIngredientHandler)

		// Tag endpoints (read)
		public.GET("/tag/:id", controller.GetTagHandler)
		public.GET("/tags", controller.GetAllTagsHandler)

		// Category endpoints (read)
		public.GET("/category/:id", controller.GetCategoryHandler)
		public.GET("/categories", controller.GetAllCategoriesHandler)

		// Authentication routes
		public.POST("/signup", controller.Signup)
		public.POST("/login", controller.Login)
		public.POST("/forgot-password", controller.ForgotPasswordHandler)
		public.POST("/reset-password", controller.ResetPasswordHandler)

		// Swagger routes
		public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	protected := r.Group("/")
	protected.Use(middleware.AuthenticatJWT())
	{
		// Recipe management
		protected.POST("/recipe", controller.PostRecipeHandler)
		protected.PUT("/recipe/:id", controller.PutRecipeUpdateHandler)
		protected.DELETE("/recipe/:id", controller.DeleteRecipeHandler)

		// Recipe tags & categories management
		protected.PUT("/recipe/:id/tags", controller.PutRecipeTagsHandler)
		protected.DELETE("/recipe/:id/tags", controller.DeleteRecipeTagsHandler)
		protected.GET("/recipe/:id/categories", controller.GetRecipeCategoriesHandler)
		protected.GET("/recipe/:id/tags", controller.GetRecipeTagsHandler)
		protected.DELETE("/recipe/:id/categories", controller.DeleteRecipeCategoriesHandler)

		// Ingredient management
		protected.POST("/ingredient", controller.PostIngredientHandler)
		protected.DELETE("/ingredient/:id", controller.DeleteIngredientHandler)

		// Comment management
		protected.POST("/comment", controller.PostCommentHandler)
		protected.DELETE("/comment/:id", controller.DeleteCommentHandler)
		protected.POST("/comment/:id/like", controller.PostCommentLikeIncHandler)
		protected.POST("/comment/:id/dislike", controller.PostCommentLikeDecHandler)

		// Favorite management
		protected.POST("/favorite", controller.PostFavoriteHandler)
		protected.DELETE("/favorite/:id", controller.DeleteFavoriteHandler)

		// Rating management
		protected.POST("/rating", controller.PostRatingHandler)
		protected.PUT("/rating/:id", controller.PutUpdateRatingHandler)
		protected.DELETE("/rating/:id", controller.DeleteRatingHandler)

		// User-specific data
		protected.GET("/user/:id/recipes", controller.GetUserRecipesHandler)
		protected.GET("/user/:id/favorites", controller.GetUserFavoritesHandler)
		protected.GET("/user/:id/ratings", controller.GetUserRatingsHandler)

		// Tag management
		protected.POST("/tag", controller.PostTagHandler)
		protected.PUT("/tag/:id", controller.PutTagHandler)
		protected.DELETE("/tag/:id", controller.DeleteTagHandler)

		// Category management
		protected.POST("/category", controller.PostCategoryHandler)
		protected.PUT("/category/:id", controller.PutCategoryHandler)
		protected.DELETE("/category/:id", controller.DeleteCategoryHandler)

		// Recipe image routes
		protected.POST("/recipe/:id/image", controller.PostUploadRecipeImageHandler)
		protected.GET("/recipe/:id/image/:imageId", controller.GetServeRecipeImageHandler)
		protected.DELETE("/recipe/:id/image/:imageId", controller.DeleteRecipeImageHandler)

		// User profile image routes
		protected.POST("/user/:id/profile-image", controller.PostUploadUserProfileImageHandler)
		protected.GET("/user/:id/profile-image/:imageId", controller.GetServeUserProfileImageHandler)
		protected.DELETE("/user/:id/profile-image/:imageId", controller.DeleteUserProfileImageHandler)

		// Generic image routes
		protected.GET("/image/:entity/:entityId/:imageId", controller.GetImageHandler)
	}
}
