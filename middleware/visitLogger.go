package middleware

import (
	"fmt"
	"net/http"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/model"
	"github.com/gin-gonic/gin"
)

func SiteVisitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		visit := model.SiteVisit{
			IPAddress: c.ClientIP(),
		}
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(uint); ok {
				visit.UserID = &uid
			}
		}
		db, err := internal.GetGormInstance()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := db.Create(&visit).Error; err != nil {
			fmt.Println(err)
		}
		c.Next()
	}
}
