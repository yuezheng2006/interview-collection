package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	WordCount int       `json:"wordCount"`
}

var inMemoryDocs = make(map[string]*Document)

func registerDocumentRoutes(g *gin.RouterGroup) {
	g.GET("/documents", func(c *gin.Context) {
		list := make([]*Document, 0, len(inMemoryDocs))
		for _, d := range inMemoryDocs {
			list = append(list, d)
		}
		c.JSON(http.StatusOK, list)
	})

	type createReq struct {
		Title string `json:"title"`
	}
	g.POST("/documents", func(c *gin.Context) {
		var req createReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		now := time.Now()
		id := now.Format("20060102150405.000000000")
		doc := &Document{
			ID:        id,
			Title:     req.Title,
			Content:   "",
			CreatedAt: now,
			UpdatedAt: now,
			WordCount: 0,
		}
		inMemoryDocs[id] = doc
		c.JSON(http.StatusOK, doc)
	})

	type updateReq struct {
		Content string `json:"content"`
		Title   string `json:"title"`
	}
	g.PUT("/documents/:id", func(c *gin.Context) {
		id := c.Param("id")
		doc, ok := inMemoryDocs[id]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
			return
		}
		var req updateReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}
		if req.Title != "" {
			doc.Title = req.Title
		}
		if req.Content != "" {
			doc.Content = req.Content
			doc.WordCount = len([]rune(req.Content))
		}
		doc.UpdatedAt = time.Now()
		c.JSON(http.StatusOK, doc)
	})

	g.DELETE("/documents/:id", func(c *gin.Context) {
		id := c.Param("id")
		delete(inMemoryDocs, id)
		c.Status(http.StatusNoContent)
	})
}
