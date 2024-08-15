package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"newsapp/services"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	v1 := r.Group("/v1")
	{
		v1.POST("/register", func(c *gin.Context) {
			RegisterUser(c, db)
		})
		v1.POST("/login", func(c *gin.Context) {
			Login(c, db)
		})
		v1.POST("/preference", func(c *gin.Context) {
			SetUserPreference(c, db)
		})
		v1.POST("/track", func(c *gin.Context) {
			TrackInteraction(c, db)
		})
		v1.GET("/news", func(c *gin.Context) {
			GetPersonalizedNews(c, db)
		})

		v1.GET("/fetch-news", func(c *gin.Context) {
			services.FetchAndStoreNewsHuggingFace(db)
			c.JSON(200, gin.H{
				"message": "News fetched and stored successfully",
			})
		})
	}
}
