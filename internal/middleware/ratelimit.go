package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	limiters = make(map[uint]*rate.Limiter)
	mu       sync.RWMutex
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerID, exists := c.Get("provider_id")
		if !exists {
			c.Next()
			return
		}

		pID := providerID.(uint)
		rpm, exists := c.Get("provider_rpm")
		if !exists || rpm.(int) <= 0 {
			c.Next()
			return
		}

		limiter := getLimiter(pID, rpm.(int))
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded for this provider"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getLimiter(id uint, rpm int) *rate.Limiter {
	mu.RLock()
	limiter, exists := limiters[id]
	mu.RUnlock()

	if exists {
		return limiter
	}

	mu.Lock()
	defer mu.Unlock()

	// Double check
	if limiter, exists := limiters[id]; exists {
		return limiter
	}

	// rpm / 60 = rps
	newLimiter := rate.NewLimiter(rate.Limit(float64(rpm)/60.0), rpm)
	limiters[id] = newLimiter
	return newLimiter
}
