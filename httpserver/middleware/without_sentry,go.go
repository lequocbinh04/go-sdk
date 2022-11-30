package middleware

import "github.com/gin-gonic/gin"

func WithoutSentry() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

			}
		}()
		c.Next()
	}
}
