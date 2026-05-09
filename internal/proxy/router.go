package proxy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/choken/llm-proxy/internal/database"
)

type Router struct {
	mu           sync.Mutex
	keyIndexMap     map[uint]int // providerID -> current key index
	channelIndexMap map[string]int // modelName -> current channel index
}

func NewRouter() *Router {
	return &Router{
		keyIndexMap:     make(map[uint]int),
		channelIndexMap: make(map[string]int),
	}
}

func (r *Router) Route(ctx context.Context, model string, policy *database.AccessToken) (ProviderAdapter, *database.Channel, *database.APIKey, string, string, error) {
	var channels []database.Channel
	query := database.DB.Preload("APIKeys").Where("is_active = ?", true)
	
	// Apply Channel Policy if set
	if policy != nil && policy.AllowedChannels != "" {
		ids := strings.Split(policy.AllowedChannels, ",")
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&channels).Error; err != nil {
		return nil, nil, nil, "", "", fmt.Errorf("failed to fetch channels: %v", err)
	}

	// 0. Check system-wide enable flags
	sysEnabled := func(chType string) bool {
		switch strings.ToLower(chType) {
		case "openai":
			return database.GetConfig("enable_openai") == "true"
		case "anthropic":
			return database.GetConfig("enable_anthropic") == "true"
		case "gemini":
			return database.GetConfig("enable_gemini") == "true"
		}
		return true
	}

	// 1. Check for explicit routing (channel/model)
	for _, ch := range channels {
		if !sysEnabled(ch.Type) {
			continue
		}
		prefix := ch.Name + "/"
		if strings.HasPrefix(model, prefix) {
			targetModel := strings.TrimPrefix(model, prefix)
			
			// Verify if the channel supports this model
			if !r.IsModelAllowed(targetModel, ch.AllowedModels, ch.DeniedModels) {
				return nil, nil, nil, "", "", fmt.Errorf("channel %s does not allow model %s", ch.Name, targetModel)
			}
			
			// Token policy still applies to the base model
			if policy != nil && !r.IsModelAllowed(targetModel, policy.AllowedModels, policy.DeniedModels) {
				return nil, nil, nil, "", "", fmt.Errorf("token does not allow model %s", targetModel)
			}

			apiKey := r.PickAPIKey(&ch)
			if apiKey == nil {
				return nil, nil, nil, "", "", fmt.Errorf("no active API key for channel %s", ch.Name)
			}
			
			return r.getAdapter(&ch), &ch, apiKey, r.getKeyHint(apiKey.KeyValue), targetModel, nil
		}
	}

	// 2. Default Load Balancing Routing (Round-Robin)
	r.mu.Lock()
	startIndex := r.channelIndexMap[model]
	r.mu.Unlock()

	numChannels := len(channels)
	for i := 0; i < numChannels; i++ {
		ch := channels[(startIndex+i)%numChannels]
		
		if !sysEnabled(ch.Type) {
			continue
		}
		// Channel's own model policy
		if !r.IsModelAllowed(model, ch.AllowedModels, ch.DeniedModels) {
			continue
		}

		// Token's model policy
		if policy != nil {
			if !r.IsModelAllowed(model, policy.AllowedModels, policy.DeniedModels) {
				continue
			}
		}
		
		apiKey := r.PickAPIKey(&ch)
		if apiKey == nil {
			continue
		}

		// Update channel index for this model
		r.mu.Lock()
		r.channelIndexMap[model] = (startIndex + i + 1) % numChannels
		r.mu.Unlock()

		return r.getAdapter(&ch), &ch, apiKey, r.getKeyHint(apiKey.KeyValue), model, nil
	}

	return nil, nil, nil, "", "", fmt.Errorf("no suitable channel found for model: %s", model)
}

func (r *Router) getAdapter(ch *database.Channel) ProviderAdapter {
	switch strings.ToLower(ch.Type) {
	case "openai":
		return &OpenAIAdapter{BaseURL: ch.BaseURL}
	case "anthropic":
		return &AnthropicAdapter{BaseURL: ch.BaseURL}
	case "gemini":
		return &GeminiAdapter{BaseURL: ch.BaseURL}
	default:
		return &OpenAIAdapter{BaseURL: ch.BaseURL}
	}
}

func (r *Router) PickAPIKey(ch *database.Channel) *database.APIKey {
	if len(ch.APIKeys) == 0 {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Filter active keys
	var activeKeys []*database.APIKey
	for i := range ch.APIKeys {
		if ch.APIKeys[i].Status == "active" {
			activeKeys = append(activeKeys, &ch.APIKeys[i])
		}
	}

	if len(activeKeys) == 0 {
		return nil
	}

	// 2. Round-Robin Selection
	index := r.keyIndexMap[ch.ID]
	selected := activeKeys[index%len(activeKeys)]
	r.keyIndexMap[ch.ID] = (index + 1) % len(activeKeys)

	now := time.Now()
	selected.LastUsedAt = &now
	return selected
}

func (r *Router) getKeyHint(key string) string {
	if len(key) <= 8 {
		return key
	}
	return "..." + key[len(key)-4:]
}

func (r *Router) IsModelAllowed(model, allowed, denied string) bool {
	model = strings.TrimSpace(model)
	
	// 1. Check Denied Models
	if denied != "" && denied != "[]" {
		for _, pattern := range strings.Split(denied, ",") {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			if r.globMatch(pattern, model) {
				return false
			}
		}
	}

	// 2. Check Allowed Models
	if allowed == "" || allowed == "*" || allowed == "[]" {
		return true
	}

	for _, pattern := range strings.Split(allowed, ",") {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if r.globMatch(pattern, model) {
			return true
		}
	}

	return false
}

// globMatch provides a simple wildcard matcher that handles slashes
func (r *Router) globMatch(pattern, text string) bool {
	if pattern == "*" {
		return true
	}
	// If no wildcards, do exact match
	if !strings.Contains(pattern, "*") {
		return pattern == text
	}
	
	// Simple prefix/suffix match for most common cases
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(text, pattern[:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(text, pattern[1:])
	}
	
	// Fallback to a simpler containment check for model names if it's just a part
	return strings.Contains(text, strings.ReplaceAll(pattern, "*", ""))
}
