package middleware

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/choken/llm-proxy/internal/database"
	"github.com/choken/llm-proxy/internal/proxy"
	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Only log LLM Proxy requests
		isProxyRequest := path == "/v1/chat/completions" || 
						 path == "/v1/messages" || 
						 strings.HasPrefix(path, "/v1/models/")

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		if isProxyRequest {
			go logRequest(c, latency)
		}
	}
}

func logRequest(c *gin.Context, latency time.Duration) {
	var atID, chID uint
	var chName, mdl, msg, kHint string

	if val, ok := c.Get("access_token_id"); ok {
		atID, _ = val.(uint)
	}
	if val, ok := c.Get("channel_id"); ok {
		chID, _ = val.(uint)
	}
	if val, ok := c.Get("channel_name"); ok {
		chName, _ = val.(string)
	}
	if val, ok := c.Get("model"); ok {
		mdl, _ = val.(string)
	}
	if val, ok := c.Get("error_message"); ok {
		msg, _ = val.(string)
	}
	if val, ok := c.Get("key_hint"); ok {
		kHint, _ = val.(string)
	}

	// Generate a simple request ID (prefix 'req-' as in user example)
	requestID := "req-" + time.Now().Format("05.000") + getString(kHint)

	var pTokens, cTokens, tTokens interface{}
	if val, ok := c.Get("prompt_tokens"); ok {
		pTokens = val
	}
	if val, ok := c.Get("completion_tokens"); ok {
		cTokens = val
	}
	if val, ok := c.Get("total_tokens"); ok {
		tTokens = val
	}

	logEntry := database.RequestLog{
		RequestID:        requestID,
		AccessTokenID:    atID,
		ChannelID:         chID,
		Provider:         chName,
		Model:             mdl,
		PromptTokens:     getInt(pTokens),
		CompletionTokens: getInt(cTokens),
		TotalTokens:       getInt(tTokens),
		KeyHint:          kHint,
		IP:               c.ClientIP(),
		StatusCode:       c.Writer.Status(),
		LatencyMS:        latency.Milliseconds(),
		ErrorMessage:     msg,
		CreatedAt:        time.Now(),
	}

	// Ensure at least 1ms for successful requests to avoid 0ms in dashboard
	if logEntry.LatencyMS == 0 && latency > 0 {
		logEntry.LatencyMS = 1
	}

	// Save to In-Memory store instead of Database
	proxy.GlobalLogStore.Add(logEntry)

	// Print to stdout in the requested JSON format
	output := map[string]interface{}{
		"timestamp":   logEntry.CreatedAt.Format(time.RFC3339),
		"request_id":  logEntry.RequestID,
		"ip":          logEntry.IP,
		"model":       logEntry.Model,
		"provider":    logEntry.Provider,
		"usage": map[string]int{
			"prompt_tokens":     logEntry.PromptTokens,
			"completion_tokens": logEntry.CompletionTokens,
			"total_tokens":      logEntry.TotalTokens,
		},
		"latency_ms":  logEntry.LatencyMS,
		"status_code": logEntry.StatusCode,
	}
	jsonBytes, _ := json.Marshal(output)
	fmt.Println(string(jsonBytes))
}

func getInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
