package handler

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func useMiddlewares(r *gin.Engine) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// auth middleware
func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "缺少认证令牌"})
			return
		}

		// 处理Bearer token格式
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:] // 去掉"Bearer "前缀
		} else {
			token = authHeader // 如果没有Bearer前缀，直接使用
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "无效的认证令牌"})
			return
		}

		username, err := parseJWT(token)
		if err != nil || username == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "无效的认证令牌"})
			return
		}
		c.Set("username", username)
		c.Next()
	}
}
