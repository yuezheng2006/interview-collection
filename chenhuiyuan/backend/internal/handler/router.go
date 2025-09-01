package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() http.Handler {
	r := gin.New()
	useMiddlewares(r)

	api := r.Group("/api")
	registerUserRoutes(api)
	protected := api.Group("")
	protected.Use(requireAuth())
	registerDocumentRoutes(protected)
	registerAIRoutes(protected)

	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })
	return r
}
