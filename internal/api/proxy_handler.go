package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/choken/llm-proxy/internal/database"
	"github.com/choken/llm-proxy/internal/proxy"
	"github.com/gin-gonic/gin"
)

func (s *Server) HandleChatCompletion(c *gin.Context) {
	if database.GetConfig("enable_openai") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "OpenAI endpoint is disabled by administrator"})
		return
	}
	var req proxy.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	s.executeRequest(c, &req, "openai")
}

func (s *Server) HandleAnthropicMessages(c *gin.Context) {
	if database.GetConfig("enable_anthropic") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anthropic endpoint is disabled by administrator"})
		return
	}
	var antReq proxy.AnthropicRequest
	if err := c.ShouldBindJSON(&antReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Anthropic request"})
		return
	}
	req := proxy.FromAnthropic(&antReq)
	s.executeRequest(c, req, "anthropic")
}

func (s *Server) HandleGeminiGenerateContent(c *gin.Context) {
	if database.GetConfig("enable_gemini") != "true" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Gemini endpoint is disabled by administrator"})
		return
	}
	modelAction := c.Param("model_action")
	model := strings.Split(modelAction, ":")[0] // Extract 'gemini-pro' from 'gemini-pro:generateContent'
	
	var gemReq proxy.GeminiRequest
	if err := c.ShouldBindJSON(&gemReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Gemini request"})
		return
	}
	req := proxy.FromGemini(&gemReq, model)
	s.executeRequest(c, req, "gemini")
}

func (s *Server) executeRequest(c *gin.Context, req *proxy.ChatCompletionRequest, sourceProtocol string) {
	var policy *database.AccessToken
	if val, exists := c.Get("access_token"); exists {
		policy = val.(*database.AccessToken)
	}

	originalModel := req.Model
	adapter, channel, apiKey, hint, targetModel, err := s.router.Route(c.Request.Context(), originalModel, policy)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	req.Model = targetModel // Use the stripped model name for upstream
	fmt.Printf("[DEBUG] Routing model: %s -> target: %s (channel: %s)\n", originalModel, targetModel, channel.Name)
	key := apiKey.KeyValue

	c.Set("channel_id", channel.ID)
	c.Set("channel_name", channel.Name)
	c.Set("model", req.Model)
	c.Set("key_hint", hint)

	if req.Stream {
		s.handleStream(c, adapter, req, apiKey, sourceProtocol)
		return
	}

	resp, err := adapter.SendChatCompletion(c.Request.Context(), req, key)
	if err != nil {
		c.Set("error_message", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("prompt_tokens", resp.Usage.PromptTokens)
	c.Set("completion_tokens", resp.Usage.CompletionTokens)
	c.Set("total_tokens", resp.Usage.TotalTokens)
	database.DB.Model(channel).Update("updated_at", time.Now())
	
	// Update API Key usage
	database.DB.Model(apiKey).Updates(map[string]interface{}{
		"total_tokens":   apiKey.TotalTokens + int64(resp.Usage.TotalTokens),
		"request_count":  apiKey.RequestCount + 1,
		"last_used_at":   time.Now(),
	})

	// Format response based on source protocol
	switch sourceProtocol {
	case "anthropic":
		c.JSON(http.StatusOK, proxy.ToAnthropic(resp))
	default:
		c.JSON(http.StatusOK, resp)
	}
}

func (s *Server) handleStream(c *gin.Context, adapter proxy.ProviderAdapter, req *proxy.ChatCompletionRequest, apiKey *database.APIKey, sourceProtocol string) {
	key := apiKey.KeyValue
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	chunkChan, errChan := adapter.StreamChatCompletion(c.Request.Context(), req, key)
	var finalUsage *proxy.Usage

	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-chunkChan:
			if !ok {
				if sourceProtocol == "anthropic" {
					c.SSEvent("message_stop", gin.H{"type": "message_stop"})
				} else {
					c.SSEvent("", "[DONE]")
				}
				return false
			}
			if chunk.Usage != nil {
				finalUsage = chunk.Usage
			}
			// Here we should ideally convert chunk to sourceProtocol format
			// For simplicity, we mostly output OpenAI format which many clients accept
			c.SSEvent("", chunk)
			return true
		case err := <-errChan:
			c.SSEvent("error", gin.H{"error": err.Error()})
			return false
		case <-c.Request.Context().Done():
			return false
		}
	})

	if finalUsage != nil {
		c.Set("prompt_tokens", finalUsage.PromptTokens)
		c.Set("completion_tokens", finalUsage.CompletionTokens)
		c.Set("total_tokens", finalUsage.TotalTokens)

		// Update API Key usage for stream
		database.DB.Model(apiKey).Updates(map[string]interface{}{
			"total_tokens":   apiKey.TotalTokens + int64(finalUsage.TotalTokens),
			"request_count":  apiKey.RequestCount + 1,
			"last_used_at":   time.Now(),
		})
	}
}
