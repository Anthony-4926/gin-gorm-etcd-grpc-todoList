package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func InitMiddleware(service []interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Keys = make(map[string]interface{})
		fmt.Println(service[0])
		c.Keys["user"] = service[0]
		c.Keys["task"] = service[1]
		c.Next()
	}
}
