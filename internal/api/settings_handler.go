package api

import (
	"net/http"

	"github.com/choken/llm-proxy/internal/database"
	"github.com/gin-gonic/gin"
)

type SettingsResponse struct {
	EnableOpenAI    bool `json:"enable_openai"`
	EnableAnthropic bool `json:"enable_anthropic"`
	EnableGemini    bool `json:"enable_gemini"`
}

func (s *Server) GetSettings(c *gin.Context) {
	resp := SettingsResponse{
		EnableOpenAI:    database.GetConfig("enable_openai") == "true",
		EnableAnthropic: database.GetConfig("enable_anthropic") == "true",
		EnableGemini:    database.GetConfig("enable_gemini") == "true",
	}
	// log.Printf("[API] GetSettings: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

func (s *Server) UpdateSettings(c *gin.Context) {
	var req SettingsResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	updateConfig := func(key string, val bool) {
		strVal := "false"
		if val {
			strVal = "true"
		}
		var item database.ConfigItem
		database.DB.Where("key = ?", key).FirstOrCreate(&item, database.ConfigItem{Key: key})
		database.DB.Model(&item).Update("value", strVal)
	}

	updateConfig("enable_openai", req.EnableOpenAI)
	updateConfig("enable_anthropic", req.EnableAnthropic)
	updateConfig("enable_gemini", req.EnableGemini)

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}
