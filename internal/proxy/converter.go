package proxy

import (
	"strings"
)

// FromAnthropic converts AnthropicRequest to internal ChatCompletionRequest
func FromAnthropic(req *AnthropicRequest) *ChatCompletionRequest {
	messages := make([]ChatMessage, 0)
	if req.System != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		messages = append(messages, ChatMessage{Role: m.Role, Content: m.Content})
	}
	return &ChatCompletionRequest{
		Model:     req.Model,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
	}
}

// FromGemini converts GeminiRequest to internal ChatCompletionRequest
func FromGemini(req *GeminiRequest, model string) *ChatCompletionRequest {
	messages := make([]ChatMessage, 0)
	if req.System != nil && len(req.System.Parts) > 0 {
		messages = append(messages, ChatMessage{Role: "system", Content: req.System.Parts[0].Text})
	}
	for _, c := range req.Contents {
		role := "user"
		if strings.ToLower(c.Role) == "model" {
			role = "assistant"
		}
		if len(c.Parts) > 0 {
			messages = append(messages, ChatMessage{Role: role, Content: c.Parts[0].Text})
		}
	}
	return &ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
}

// ToAnthropic converts internal ChatCompletionResponse to Anthropic format
// (Simplified, can be expanded as needed)
func ToAnthropic(resp *ChatCompletionResponse) map[string]interface{} {
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}
	return map[string]interface{}{
		"id":      resp.ID,
		"type":    "message",
		"role":    "assistant",
		"content": []map[string]interface{}{{"type": "text", "text": content}},
		"model":   resp.Model,
		"usage": map[string]interface{}{
			"input_tokens":  resp.Usage.PromptTokens,
			"output_tokens": resp.Usage.CompletionTokens,
		},
	}
}
