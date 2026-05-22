package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ClientIP 获取访客真实 IP（优先代理头）
func ClientIP(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	return c.ClientIP()
}
