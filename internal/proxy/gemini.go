package proxy

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiAdapter struct {
	BaseURL string
}

func (a *GeminiAdapter) SendChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (*ChatCompletionResponse, error) {
	config := &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	}
	// If custom base URL is provided, it might need more complex config depending on the SDK's support
	// For now we assume standard Gemini API

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %v", err)
	}

	var contents []*genai.Content
	var systemInstruction *genai.Content

	for _, m := range req.Messages {
		role := "user"
		if strings.ToLower(m.Role) == "assistant" || strings.ToLower(m.Role) == "model" {
			role = "model"
		}

		if strings.ToLower(m.Role) == "system" {
			systemInstruction = &genai.Content{
				Parts: []*genai.Part{{Text: m.Content}},
			}
			continue
		}

		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: m.Content}},
		})
	}

	generateConfig := &genai.GenerateContentConfig{
		SystemInstruction: systemInstruction,
	}
	if req.Temperature > 0 {
		generateConfig.Temperature = genai.Ptr(float32(req.Temperature))
	}
	if req.MaxTokens > 0 {
		generateConfig.MaxOutputTokens = int32(req.MaxTokens)
	}

	resp, err := client.Models.GenerateContent(ctx, req.Model, contents, generateConfig)
	if err != nil {
		return nil, fmt.Errorf("gemini error: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	content := resp.Candidates[0].Content.Parts[0].Text

	return &ChatCompletionResponse{
		ID:      "gemini-resp",
		Object:  "chat.completion",
		Created: 0,
		Model:   req.Model,
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: string(resp.Candidates[0].FinishReason),
			},
		},
		Usage: Usage{
			PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
		},
	}, nil
}
func (a *GeminiAdapter) StreamChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (<-chan *StreamResponseChunk, <-chan error) {
	chunkChan := make(chan *StreamResponseChunk)
	errChan := make(chan error, 1)

	config := &genai.ClientConfig{APIKey: apiKey, Backend: genai.BackendGeminiAPI}
	client, _ := genai.NewClient(ctx, config)

	var contents []*genai.Content
	var systemInstruction *genai.Content
	for _, m := range req.Messages {
		role := "user"
		if strings.ToLower(m.Role) == "assistant" || strings.ToLower(m.Role) == "model" {
			role = "model"
		}
		if strings.ToLower(m.Role) == "system" {
			systemInstruction = &genai.Content{Parts: []*genai.Part{{Text: m.Content}}}
			continue
		}
		contents = append(contents, &genai.Content{Role: role, Parts: []*genai.Part{{Text: m.Content}}})
	}

	generateConfig := &genai.GenerateContentConfig{SystemInstruction: systemInstruction}
	if req.Temperature > 0 {
		generateConfig.Temperature = genai.Ptr(float32(req.Temperature))
	}
	if req.MaxTokens > 0 {
		generateConfig.MaxOutputTokens = int32(req.MaxTokens)
	}

	go func() {
		defer close(chunkChan)
		iter := client.Models.GenerateContentStream(ctx, req.Model, contents, generateConfig)
		
		for resp, err := range iter {
			if err != nil {
				errChan <- err
				return
			}

			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				continue
			}

			var usage *Usage
			if resp.UsageMetadata != nil {
				usage = &Usage{
					PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
					CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
					TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
				}
			}

			chunkChan <- &StreamResponseChunk{
				ID:      "gemini-stream",
				Object:  "chat.completion.chunk",
				Created: 0,
				Model:   req.Model,
				Choices: []StreamChoice{{
					Index: 0,
					Delta: ChatDelta{Content: resp.Candidates[0].Content.Parts[0].Text},
					FinishReason: string(resp.Candidates[0].FinishReason),
				}},
				Usage: usage,
			}
		}
	}()

	return chunkChan, errChan
}
