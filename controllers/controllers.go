package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"newsapp/common"
	"newsapp/models"
	"newsapp/services"
)

func RegisterUser(c *gin.Context, db *gorm.DB) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		common.HandleError(c, err, http.StatusBadRequest)
		return
	}

	// Hash the password before storing
	hashedPassword, err := common.HashPassword(user.Password)
	if err != nil {
		common.HandleError(c, err, http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	db.Create(&user)
	c.JSON(http.StatusOK, user)
}

func Login(c *gin.Context, db *gorm.DB) {
	var user models.User
	var requestUser models.User

	if err := c.ShouldBindJSON(&requestUser); err != nil {
		common.HandleError(c, err, http.StatusBadRequest)
		return
	}

	if err := db.Where("email = ?", requestUser.Email).First(&user).Error; err != nil {
		common.HandleError(c, err, http.StatusUnauthorized)
		return
	}

	if !common.CheckPasswordHash(requestUser.Password, user.Password) {
		err := fmt.Errorf("invalid password")
		common.HandleError(c, err, http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := common.GenerateJWT(user)
	if err != nil {
		common.HandleError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func SetUserPreference(c *gin.Context, db *gorm.DB) {
	var pref models.UserPreference
	if err := c.ShouldBindJSON(&pref); err != nil {
		common.HandleError(c, err, http.StatusBadRequest)
		return
	}
	db.Create(&pref)
	c.JSON(http.StatusOK, pref)
}

func TrackInteraction(c *gin.Context, db *gorm.DB) {
	var interaction models.UserInteraction
	if err := c.ShouldBindJSON(&interaction); err != nil {
		common.HandleError(c, err, http.StatusBadRequest)
		return
	}
	db.Create(&interaction)
	c.JSON(http.StatusOK, interaction)
}

func GetPersonalizedNews(c *gin.Context, db *gorm.DB) {
	var prefs []models.UserPreference
	var interactions []models.UserInteraction
	var news []models.News
	userID := c.Query("user_id")

	db.Where("user_id = ?", userID).Find(&prefs)
	db.Where("user_id = ?", userID).Find(&interactions)

	prefCategories := []string{}
	for _, pref := range prefs {
		prefCategories = append(prefCategories, pref.Category)
	}

	// Generate personalized news
	news = services.GeneratePersonalizedNewsHuggingFace(db, userID, prefCategories, interactions)

	c.JSON(http.StatusOK, news)
}
