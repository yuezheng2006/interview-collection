package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

var users = map[string]*user{}

func registerUserRoutes(g *gin.RouterGroup) {
	g.POST("/auth/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		if _, ok := users[req.Username]; ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户已存在"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		u := &user{ID: time.Now().Format("20060102150405"), Username: req.Username, Password: string(hash)}
		users[req.Username] = u
		c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
	})

	g.POST("/auth/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		u, ok := users[req.Username]
		if !ok || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		token, err := issueJWT(u.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token, "message": "登录成功"})
	})
}
