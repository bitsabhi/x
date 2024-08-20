package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"newsapp/middleware"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/v1")
	{
		// Public routes (no JWT required)
		v1.POST("/register", func(c *gin.Context) {
			RegisterUser(c, db)
		})
		v1.POST("/login", func(c *gin.Context) {
			Login(c, db)
		})

		// Protected routes (JWT required)
		v1.Use(middleware.JWTAuthMiddleware()).POST("/preference", func(c *gin.Context) {
			SetUserPreference(c, db)
		})
		v1.Use(middleware.JWTAuthMiddleware()).POST("/track", func(c *gin.Context) {
			TrackInteraction(c, db)
		})
		v1.Use(middleware.JWTAuthMiddleware()).GET("/news", func(c *gin.Context) {
			GetPersonalizedNews(c, db)
		})
	}
}
