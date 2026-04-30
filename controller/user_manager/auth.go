package user_manager

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"

	"github.com/gin-gonic/gin"
)

// Auth 校验外部用户管理接口的固定授权码。
func Auth() func(c *gin.Context) {
	return func(c *gin.Context) {
		if common.UserManagerAuthKey == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "用户管理接口未配置授权码",
			})
			c.Abort()
			return
		}
		key := strings.TrimSpace(c.GetHeader("X-User-Manager-Key"))
		if key == "" {
			key = strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(key, "Bearer ") || strings.HasPrefix(key, "bearer ") {
				key = strings.TrimSpace(key[7:])
			}
		}
		if key == "" || key != common.UserManagerAuthKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "用户管理接口授权码无效",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
