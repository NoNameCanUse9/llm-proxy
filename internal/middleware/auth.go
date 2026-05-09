package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/choken/llm-proxy/internal/database"
	"github.com/gin-gonic/gin"
)

func SKAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		token := parts[1]
		hash := hashToken(token)

		var accessToken database.AccessToken
		err := database.DB.Where("token_hash = ?", hash).First(&accessToken).Error
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		if !accessToken.IsActive {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is inactive"})
			c.Abort()
			return
		}

		// 1. IP Validation
		clientIP := c.ClientIP()
		if !isIPAllowed(clientIP, accessToken.AllowedIPs, accessToken.DeniedIPs) {
			c.JSON(http.StatusForbidden, gin.H{"error": "IP address not allowed"})
			c.Abort()
			return
		}

		// 2. RPM Rate Limiting
		rpm := 0
		if accessToken.RPM != nil {
			rpm = *accessToken.RPM
		}
		if !checkRateLimit(accessToken.ID, rpm) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded (RPM)"})
			c.Abort()
			return
		}

		// Set access token in context for policy enforcement and logging
		c.Set("access_token_id", accessToken.ID)
		c.Set("access_token", &accessToken)
		c.Next()
	}
}

func isIPAllowed(ip, allowed, denied string) bool {
	// Simple comma-separated check for now. CIDR can be added later if needed.
	if denied != "" {
		ips := strings.Split(denied, ",")
		for _, v := range ips {
			if strings.TrimSpace(v) == ip {
				return false
			}
		}
	}
	if allowed != "" {
		ips := strings.Split(allowed, ",")
		for _, v := range ips {
			if strings.TrimSpace(v) == ip || strings.TrimSpace(v) == "*" {
				return true
			}
		}
		return false // Not in allowed list
	}
	return true
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}


var (
	rateLimitMap = make(map[uint]*tokenLimiter)
	rateLimitMu  sync.Mutex
)

type tokenLimiter struct {
	count     int
	lastReset time.Time
	mu        sync.Mutex
}

func checkRateLimit(tokenID uint, rpm int) bool {
	if rpm <= 0 {
		return true // No limit
	}

	rateLimitMu.Lock()
	limiter, exists := rateLimitMap[tokenID]
	if !exists {
		limiter = &tokenLimiter{lastReset: time.Now()}
		rateLimitMap[tokenID] = limiter
	}
	rateLimitMu.Unlock()

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if time.Since(limiter.lastReset) > time.Minute {
		limiter.count = 0
		limiter.lastReset = time.Now()
	}

	if limiter.count >= rpm {
		return false
	}

	limiter.count++
	return true
}
