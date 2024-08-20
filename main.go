package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"newsapp/controllers"
	"newsapp/middleware"
	"newsapp/models"
	"newsapp/services"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Debugging: Print the environment variables
	fmt.Println("HUGGINGFACE_API_KEY:", os.Getenv("HUGGINGFACE_API_KEY"))
	fmt.Println("NEWSAPI_KEY:", os.Getenv("NEWSAPI_KEY"))
}

var db *gorm.DB

// Handler function to manually trigger news fetching
func FetchNewsHandler(c *gin.Context) {
	services.FetchAndStoreNewsHuggingFace(db)
	c.JSON(http.StatusOK, gin.H{"status": "News fetched and stored successfully"})
}

func main() {
	// todo: gin.SetMode(gin.ReleaseMode)  // Set Gin to release mode for production

	var err error
	// Open a connection to the SQLite database using the newer GORM package
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Migrate the schema (create the necessary tables)
	autoMigrateTables()

	// Fetch and store news articles at startup
	services.FetchAndStoreNewsHuggingFace(db)

	// Initialize Gin router for handling HTTP requests
	r := gin.Default()

	// Apply middleware for rate limiting
	r.Use(middleware.RateLimitMiddleware())

	// Enable CloudWatch for logging in production environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.DisableConsoleColor()
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	} else {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	// Register all routes from routes.go
	controllers.RegisterRoutes(r, db)

	// Define API routes
	r.GET("/", func(c *gin.Context) {
		c.String(200, "Welcome to the news app!")
	})

	// Add the route for manually fetching news
	r.GET("/fetch-news", FetchNewsHandler)

	fmt.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}

func autoMigrateTables() {
	// todo
	db.AutoMigrate(&models.News{})            // Add other models here
	db.AutoMigrate(&models.User{})            // Example: User model
	db.AutoMigrate(&models.UserPreference{})  // Example: Preference model
	db.AutoMigrate(&models.UserInteraction{}) // Example: Interaction model
}
