package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"time"
)

func RateLimitMiddleware() gin.HandlerFunc {
	// Set rate limit to control traffic to the API
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  20,
	}

	// In-memory store for rate limiting (consider external store for production)
	store := memory.NewStore()

	instance := limiter.New(store, rate)

	// Create the middleware
	return ginlimiter.NewMiddleware(instance)
}
