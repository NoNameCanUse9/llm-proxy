package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/choken/llm-proxy/internal/config"
	"github.com/choken/llm-proxy/internal/database"
	"github.com/choken/llm-proxy/internal/middleware"
	"github.com/choken/llm-proxy/internal/proxy"
	"github.com/gin-gonic/gin"

	_ "github.com/choken/llm-proxy/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router    *proxy.Router
	StartTime time.Time
}

func NewServer() *Server {
	return &Server{
		router:    proxy.NewRouter(),
		StartTime: time.Now(),
	}
}

func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

func (s *Server) Start() error {
	r := gin.Default()

	// Global middleware
	r.Use(middleware.LoggerMiddleware())
	r.Use(NoCacheMiddleware())

	// Static files for Frontend
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	
	// Root route serves index.html
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	// Public routes
	r.POST("/auth/login", s.HandleAdminLogin)
	r.GET("/status", s.HandleStatus)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Admin routes (JWT)
	admin := r.Group("/admin")
	admin.Use(middleware.JWTMiddleware())
	{
		// Channels
		admin.GET("/channels", s.ListChannels)
		admin.POST("/channels", s.CreateChannel)
		admin.PUT("/channels/:id", s.UpdateChannel)
		admin.DELETE("/channels/:id", s.DeleteChannel)
		admin.Match([]string{"GET", "POST"}, "/channels/fetch-models", s.FetchUpstreamModels)

		// Client Tokens
		admin.GET("/tokens", s.ListAccessTokens)
		admin.POST("/tokens", s.CreateAccessToken)
		admin.PUT("/tokens/:id", s.UpdateAccessToken)
		admin.DELETE("/tokens/:id", s.DeleteAccessToken)

		// Logs
		admin.GET("/logs", s.ListLogs)

		// User
		admin.POST("/user/password", s.ChangePassword)
		admin.POST("/user/username", s.HandleUpdateUsername)

		// Settings
		admin.PUT("/settings", s.UpdateSettings)
	}

	// PUBLIC SETTINGS ROUTE FOR DEBUGGING
	r.GET("/admin/settings", s.GetSettings)

	// Proxy routes (Bearer sk-xxx)
	v1 := r.Group("/v1")
	v1.Use(middleware.SKAuthMiddleware())
	{
		v1.POST("/chat/completions", s.HandleChatCompletion)
		v1.POST("/messages", s.HandleAnthropicMessages)
		v1.POST("/models/:model_action", s.HandleGeminiGenerateContent)
		v1.GET("/models", s.HandleListModels)
	}

	// SPA Support: Catch-all for frontend routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// If the request starts with API prefixes but wasn't matched, return 404
		if strings.HasPrefix(path, "/v1") || strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/auth") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API route not found"})
			return
		}

		// Development Proxy
		if os.Getenv("DEV_MODE") == "true" {
			target, _ := url.Parse("http://localhost:5173")
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Otherwise, serve index.html for SPA routing (Production)
		c.File("./frontend/dist/index.html")
	})

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Port)
	fmt.Printf("Server starting on %s\n", addr)
	return r.Run(addr)
}

// HandleListModels returns a list of available models for the given access token
// @Summary List Models
// @Description Returns a list of models allowed for the current access token.
// @Tags Client API
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/models [get]
// @Security ApiKeyAuth
func (s *Server) HandleListModels(c *gin.Context) {
	accessToken, exists := c.Get("access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	policy := accessToken.(*database.AccessToken)

	var channels []database.Channel
	query := database.DB.Where("is_active = ?", true)

	// Apply Channel Policy if set
	if policy.AllowedChannels != "" {
		ids := strings.Split(policy.AllowedChannels, ",")
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}

	// Aggregate and filter models
	modelMap := make(map[string]bool)
	for _, ch := range channels {
		if ch.AllowedModels == "" || ch.AllowedModels == "*" {
			// If channel allows everything, we might want to return some defaults or skip
			// For now, let's look at what the token allows if channel is open
			if policy.AllowedModels != "" && policy.AllowedModels != "*" {
				for _, m := range strings.Split(policy.AllowedModels, ",") {
					m = strings.TrimSpace(m)
					if m != "" && !strings.Contains(m, "*") {
						modelMap[m] = true
					}
				}
			}
			continue
		}

		models := strings.Split(ch.AllowedModels, ",")
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m == "" || m == "*" {
				continue
			}
			
			// Filter by Token Policy
			if s.router.IsModelAllowed(m, policy.AllowedModels, policy.DeniedModels) {
				// 1. Add normal model (for load balancing)
				modelMap[m] = true
				// 2. Add prefixed model (for explicit routing)
				modelMap[ch.Name+"/"+m] = true
			}
		}
	}

	type ModelItem struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	var responseData []ModelItem
	now := time.Now().Unix()
	for m := range modelMap {
		responseData = append(responseData, ModelItem{
			ID:      m,
			Object:  "model",
			Created: now,
			OwnedBy: "llm-proxy",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   responseData,
	})
}
