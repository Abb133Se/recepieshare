package middleware

import (
	"net/http"
	"strings"

	"github.com/Abb133Se/recepieshare/controller"
	"github.com/Abb133Se/recepieshare/token"
	"github.com/gin-gonic/gin"
)

func AuthenticatJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization token required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		claims, err := token.VerifyToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("claims", claims)

		if idFloat, ok := claims["sub"].(float64); ok {
			c.Set("userID", uint(idFloat))
		} else {
			c.JSON(http.StatusUnauthorized, controller.SimpleMessageResponse{Message: "user id not found in token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func ExtractUserFromToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr := parts[1]
				claims, err := token.VerifyToken(tokenStr)
				if err == nil {
					if idFloat, ok := claims["sub"].(float64); ok {
						c.Set("userID", uint(idFloat))
					}
				}
			}
		}
		c.Next()
	}
}
