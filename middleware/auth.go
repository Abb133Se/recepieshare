package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Abb133Se/recepieshare/token"
	"github.com/Abb133Se/recepieshare/utils"
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in token claims"})
			c.Abort()
			return
		}
		if role, ok := claims["role"]; ok {
			c.Set("role", role)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found in token claims"})
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

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

// utils.go
func GetEffectiveUserID(c *gin.Context) (uint, error) {
	adminRole := c.GetString("role")
	paramUserID := c.Param("userID") // optional query parameter
	currentUserID := c.GetUint("userID")

	if adminRole == "admin" && paramUserID != "" {
		uid, err := utils.ValidateEntityID(paramUserID)
		if err != nil {
			return 0, fmt.Errorf("invalid user_id parameter")
		}
		return uint(uid), nil
	}

	// Non-admin or no param: only own userID
	if currentUserID == 0 {
		return 0, fmt.Errorf("unauthorized")
	}
	return currentUserID, nil
}
