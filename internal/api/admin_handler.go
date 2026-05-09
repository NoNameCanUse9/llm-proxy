package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/choken/llm-proxy/internal/database"
	"github.com/choken/llm-proxy/internal/proxy"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleAdminLogin handles administrator login
// @Summary Admin Login
// @Description Login with username and password to receive a JWT token.
// @Tags Admin API
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login Request"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (s *Server) HandleAdminLogin(c *gin.Context) {
	var loginReq LoginRequest

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user database.User
	if err := database.DB.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	secret := database.GetConfig("jwt_secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Channels
// ListChannels returns all configured channels
// @Summary List Channels
// @Description Retrieve a list of all configured upstream channels.
// @Tags Admin API
// @Produce json
// @Success 200 {array} database.Channel
// @Router /admin/channels [get]
// @Security ApiKeyAuth
func (s *Server) ListChannels(c *gin.Context) {
	var channels []database.Channel
	database.DB.Preload("APIKeys").Find(&channels)
	c.JSON(http.StatusOK, channels)
}

type ChannelRequest struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	BaseURL       string   `json:"base_url"`
	APIKeys       []string `json:"api_keys"`
	AllowedModels string   `json:"allowed_models"`
	DeniedModels  string   `json:"denied_models"`
	IsActive      bool     `json:"is_active"`
	RPM           int      `json:"rpm"`
}

// CreateChannel creates a new upstream channel
// @Summary Create Channel
// @Description Create a new upstream channel with API keys and model policies.
// @Tags Admin API
// @Accept json
// @Produce json
// @Param request body ChannelRequest true "Channel Request"
// @Success 201 {object} database.Channel
// @Router /admin/channels [post]
// @Security ApiKeyAuth
func (s *Server) CreateChannel(c *gin.Context) {
	var req ChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	channel := database.Channel{
		Name:          req.Name,
		Type:          req.Type,
		BaseURL:       req.BaseURL,
		AllowedModels: req.AllowedModels,
		DeniedModels:  req.DeniedModels,
		IsActive:      req.IsActive,
		RPM:           req.RPM,
	}

	for _, k := range req.APIKeys {
		channel.APIKeys = append(channel.APIKeys, database.APIKey{KeyValue: k})
	}

	if err := database.DB.Create(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, channel)
}

func (s *Server) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	var req ChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var channel database.Channel
	if err := database.DB.Preload("APIKeys").First(&channel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	// Update fields
	channel.Name = req.Name
	channel.Type = req.Type
	channel.BaseURL = req.BaseURL
	channel.AllowedModels = req.AllowedModels
	channel.DeniedModels = req.DeniedModels
	channel.IsActive = req.IsActive
	channel.RPM = req.RPM

	// 1. Update channel basic info
	if err := database.DB.Save(&channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel: " + err.Error()})
		return
	}

	// 2. Refresh API Keys (Manual approach to be 100% safe)
	// Delete old keys (Physical delete to be sure)
	database.DB.Unscoped().Where("provider_id = ?", channel.ID).Delete(&database.APIKey{})
	
	// Create new keys
	newKeys := convertToAPIKeys(req.APIKeys)
	for i := range newKeys {
		newKeys[i].ProviderID = channel.ID
		if err := database.DB.Create(&newKeys[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key: " + err.Error()})
			return
		}
	}

	// Reload channel with new keys to return correct response
	database.DB.Preload("APIKeys").First(&channel, channel.ID)
	c.JSON(http.StatusOK, channel)
}

func convertToAPIKeys(keys []string) []database.APIKey {
	var result []database.APIKey
	for _, k := range keys {
		result = append(result, database.APIKey{
			KeyValue: k,
			Status:   "active",
		})
	}
	return result
}

func (s *Server) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&database.Channel{}, id)
	c.JSON(http.StatusNoContent, nil)
}

// AccessTokenRequest defines the incoming payload from frontend
type AccessTokenRequest struct {
	Name            string `json:"name"`
	IsActive        bool   `json:"is_active"`
	ExpiresAt       string `json:"expires_at"` // "permanent", "7d", "30d", "180d", "365d"
	AllowedChannels string `json:"allowed_channels"`
	AllowedModels   string `json:"allowed_models"`
	DeniedModels    string `json:"denied_models"`
	AllowedIPs      string `json:"allowed_ips"`
	DeniedIPs       string `json:"denied_ips"`
	RPM             *int   `json:"rpm"`
}

func parseExpiry(s string) *time.Time {
	if s == "permanent" || s == "" {
		return nil
	}
	var duration time.Duration
	switch s {
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	case "180d":
		duration = 180 * 24 * time.Hour
	case "365d":
		duration = 365 * 24 * time.Hour
	default:
		return nil
	}
	t := time.Now().Add(duration)
	return &t
}

// Access Tokens (sk-xxx)
// ListAccessTokens returns all client access tokens
// @Summary List Access Tokens
// @Description Retrieve a list of all client access tokens (sk-xxx).
// @Tags Admin API
// @Produce json
// @Success 200 {array} database.AccessToken
// @Router /admin/tokens [get]
// @Security ApiKeyAuth
func (s *Server) ListAccessTokens(c *gin.Context) {
	var tokens []database.AccessToken
	database.DB.Find(&tokens)
	c.JSON(http.StatusOK, tokens)
}

func (s *Server) CreateAccessToken(c *gin.Context) {
	var req AccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	rawToken := "sk-" + generateRandomString(32)
	hash := hashToken(rawToken)

	token := database.AccessToken{
		TokenHash:       hash,
		Token:           rawToken,
		Name:            req.Name,
		IsActive:        true,
		ExpiresAt:       parseExpiry(req.ExpiresAt),
		AllowedChannels: req.AllowedChannels,
		AllowedModels:   req.AllowedModels,
		DeniedModels:    req.DeniedModels,
		AllowedIPs:      req.AllowedIPs,
		DeniedIPs:       req.DeniedIPs,
		RPM:             req.RPM,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// If RPM is 0 or nil, set to nil for database (Unlimited)
	if token.RPM != nil && *token.RPM <= 0 {
		token.RPM = nil
	}
	
	if err := database.DB.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"name":  token.Name,
		"token": rawToken,
		"info":  "This token will only be shown once!",
	})
}

func (s *Server) UpdateAccessToken(c *gin.Context) {
	id := c.Param("id")
	var token database.AccessToken
	if err := database.DB.First(&token, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	var req AccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Update fields
	token.Name = req.Name
	token.IsActive = req.IsActive
	token.ExpiresAt = parseExpiry(req.ExpiresAt)
	token.AllowedChannels = req.AllowedChannels
	token.AllowedModels = req.AllowedModels
	token.DeniedModels = req.DeniedModels
	token.AllowedIPs = req.AllowedIPs
	token.DeniedIPs = req.DeniedIPs
	token.RPM = req.RPM
	token.UpdatedAt = time.Now()

	if err := database.DB.Save(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
}

func (s *Server) DeleteAccessToken(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&database.AccessToken{}, id)
	c.JSON(http.StatusNoContent, nil)
}

// Logs
// ListLogs returns request logs
// @Summary List Logs
// @Description Retrieve request logs with optional filtering.
// @Tags Admin API
// @Produce json
// @Param model query string false "Filter by model name"
// @Param provider query string false "Filter by provider"
// @Param status_code query int false "Filter by status code"
// @Success 200 {array} database.RequestLog
// @Router /admin/logs [get]
// @Security ApiKeyAuth
func (s *Server) ListLogs(c *gin.Context) {
	allLogs := proxy.GlobalLogStore.GetAll()
	filtered := make([]database.RequestLog, 0)

	modelFilter := strings.ToLower(c.Query("model"))
	providerFilter := strings.ToLower(c.Query("provider"))
	ipFilter := strings.ToLower(c.Query("ip"))
	statusFilter := c.Query("status_code")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	for _, raw := range allLogs {
		log, ok := raw.(database.RequestLog)
		if !ok {
			continue
		}

		// Apply filters
		if modelFilter != "" && !strings.Contains(strings.ToLower(log.Model), modelFilter) {
			continue
		}
		if providerFilter != "" && !strings.Contains(strings.ToLower(log.Provider), providerFilter) {
			continue
		}
		if ipFilter != "" && !strings.Contains(strings.ToLower(log.IP), ipFilter) {
			continue
		}
		if statusFilter != "" && strconv.Itoa(log.StatusCode) != statusFilter {
			continue
		}
		
		// Date range filtering
		if startDate != "" {
			start, _ := time.Parse("2006-01-02", startDate)
			if log.CreatedAt.Before(start) {
				continue
			}
		}
		if endDate != "" {
			end, _ := time.Parse("2006-01-02", endDate)
			// End of the day
			end = end.Add(24 * time.Hour)
			if log.CreatedAt.After(end) {
				continue
			}
		}

		filtered = append(filtered, log)
	}

	// Basic pagination for memory slice
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 50
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(filtered) {
		c.JSON(http.StatusOK, []database.RequestLog{})
		return
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	c.JSON(http.StatusOK, filtered[start:end])
}

// User Password
func (s *Server) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.GetString("username")
	var user database.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	database.DB.Model(&user).Update("password", string(hashedPassword))
	
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (s *Server) HandleUpdateUsername(c *gin.Context) {
	var req struct {
		NewUsername string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldUsername := c.GetString("username")
	if err := database.DB.Model(&database.User{}).Where("username = ?", oldUsername).Update("username", req.NewUsername).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Username updated successfully. Please re-login."})
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Server) FetchUpstreamModels(c *gin.Context) {
	var req struct {
		BaseURL string `json:"base_url" form:"base_url"`
		APIKey  string `json:"api_key" form:"api_key"`
	}

	// Try to bind from JSON first, then from Query parameters
	if err := c.ShouldBindJSON(&req); err != nil {
		if err := c.ShouldBindQuery(&req); err != nil {
			// If both fail, we might still have data if it was a simple form
		}
	}

	if req.BaseURL == "" || req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Base URL and API Key are required"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	
	// Use the first key if multiple keys are provided
	apiKey := req.APIKey
	if lines := strings.Split(apiKey, "\n"); len(lines) > 0 {
		apiKey = strings.TrimSpace(lines[0])
	}
	
	// CRITICAL: Clean the ENTIRE string, not just ends!
	apiKey = strings.Map(func(r rune) rune {
		if r > 32 && r < 127 {
			return r
		}
		return -1 // Drop this character
	}, apiKey)

	// Robust URL joining
	base := strings.TrimRight(req.BaseURL, "/")
	if !strings.HasSuffix(base, "/models") {
		base = base + "/models"
	}
	
	request, err := http.NewRequest("GET", base, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败: " + err.Error()})
		return
	}

	// Try setting header directly in the map to bypass some Go internal validation if any
	request.Header["Authorization"] = []string{"Bearer " + apiKey}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "LLM-Proxy/1.0")

	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("[Error] Network failure: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "网络连接失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取响应失败: " + err.Error()})
		return
	}

	fmt.Printf("[Debug] Upstream Status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[Debug] Upstream Error Body: %s\n", string(bodyBytes))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("上游返回错误 %d: %s", resp.StatusCode, string(bodyBytes))})
		return
	}

	// Try standard OpenAI format first
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(bodyBytes, &result); err == nil && len(result.Data) > 0 {
		var models []string
		for _, m := range result.Data {
			models = append(models, m.ID)
		}
		c.JSON(http.StatusOK, gin.H{"models": strings.Join(models, ",")})
		return
	}

	// Try simple array format: ["model1", "model2"]
	var listResult []string
	if err := json.Unmarshal(bodyBytes, &listResult); err == nil && len(listResult) > 0 {
		c.JSON(http.StatusOK, gin.H{"models": strings.Join(listResult, ",")})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析模型列表格式: " + string(bodyBytes)})
}
func (s *Server) HandleStatus(c *gin.Context) {
	logs := proxy.GlobalLogStore.GetAll()
	
	stats := struct {
		TotalRequests  int     `json:"total_requests"`
		TotalTokens    int     `json:"total_tokens"`
		AvgLatencyMS   float64 `json:"avg_latency_ms"`
		FailedRequests int     `json:"failed_requests"`
		UptimeSeconds  float64 `json:"uptime_seconds"`
		Providers      map[string]struct {
			Requests   int     `json:"requests"`
			AvgLatency float64 `json:"avg_latency"`
			Tokens     int     `json:"tokens"`
		} `json:"providers"`
	}{
		Providers: make(map[string]struct {
			Requests   int     `json:"requests"`
			AvgLatency float64 `json:"avg_latency"`
			Tokens     int     `json:"tokens"`
		}),
	}

	stats.UptimeSeconds = time.Since(s.StartTime).Seconds()

	if len(logs) == 0 {
		c.JSON(http.StatusOK, stats)
		return
	}

	totalLatency := int64(0)
	providerLatency := make(map[string]int64)
	providerRequests := make(map[string]int)
	providerTokens := make(map[string]int)

	for _, l := range logs {
		log, ok := l.(database.RequestLog)
		if !ok {
			continue
		}

		stats.TotalRequests++
		stats.TotalTokens += log.TotalTokens
		totalLatency += log.LatencyMS

		if log.StatusCode >= 400 {
			stats.FailedRequests++
		}

		if log.Provider != "" {
			providerRequests[log.Provider]++
			providerLatency[log.Provider] += log.LatencyMS
			providerTokens[log.Provider] += log.TotalTokens
		}
	}

	if stats.TotalRequests > 0 {
		stats.AvgLatencyMS = float64(totalLatency) / float64(stats.TotalRequests)
	}

	for p, count := range providerRequests {
		pStats := stats.Providers[p]
		pStats.Requests = count
		pStats.Tokens = providerTokens[p]
		if count > 0 {
			pStats.AvgLatency = float64(providerLatency[p]) / float64(count)
		}
		stats.Providers[p] = pStats
	}

	c.JSON(http.StatusOK, stats)
}
