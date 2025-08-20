package middleware

import (
	"github.com/Abb133Se/recepieshare/messages"
	"github.com/gin-gonic/gin"
)

func SetLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		// fmt.Println(lang)
		if lang == "" {
			lang = "fa"
		}
		messages.SetLang(lang)
		c.Next()
	}
}
