package middleware

import "github.com/gin-gonic/gin"

// OfficeToolsDataPolicy 标记办公工具接口响应为临时数据（不落库、不长期缓存）
func OfficeToolsDataPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Office-Data-Policy", "ephemeral; retention=24h")
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
