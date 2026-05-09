package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicAdapter struct {
	BaseURL string
}

func (a *AnthropicAdapter) SendChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (*ChatCompletionResponse, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if a.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(a.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	var systemPrompt string
	var messages []anthropic.MessageParam

	for _, m := range req.Messages {
		if strings.ToLower(m.Role) == "system" {
			systemPrompt = m.Content
			continue
		}

		role := anthropic.MessageParamRoleUser
		if strings.ToLower(m.Role) == "assistant" {
			role = anthropic.MessageParamRoleAssistant
		}

		messages = append(messages, anthropic.MessageParam{
			Role:    role,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(m.Content)},
		})
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}
	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(float64(req.Temperature))
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic error: %v", err)
	}

	var content string
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}

	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: 0, // Anthropic doesn't provide this in the same way
		Model:   string(resp.Model),
		Choices: []ChatCompletionChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: string(resp.StopReason),
			},
		},
		Usage: Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}, nil
}
func (a *AnthropicAdapter) StreamChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (<-chan *StreamResponseChunk, <-chan error) {
	chunkChan := make(chan *StreamResponseChunk)
	errChan := make(chan error, 1)

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if a.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(a.BaseURL))
	}
	client := anthropic.NewClient(opts...)

	var systemPrompt string
	var messages []anthropic.MessageParam
	for _, m := range req.Messages {
		if strings.ToLower(m.Role) == "system" {
			systemPrompt = m.Content
			continue
		}
		role := anthropic.MessageParamRoleUser
		if strings.ToLower(m.Role) == "assistant" {
			role = anthropic.MessageParamRoleAssistant
		}
		messages = append(messages, anthropic.MessageParam{
			Role:    role,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(m.Content)},
		})
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	go func() {
		defer close(chunkChan)
		stream := client.Messages.NewStreaming(ctx, params)
		
		id := "anthropic-" + req.Model
		for stream.Next() {
			event := stream.Current()
			var delta string
			var finishReason string
			
			// Handle different event types manually using string comparison
			if string(event.Type) == "content_block_delta" {
				delta = event.Delta.Text
			} else if string(event.Type) == "message_delta" {
				finishReason = string(event.Delta.StopReason)
			}

			if delta == "" && finishReason == "" {
				continue
			}

			chunkChan <- &StreamResponseChunk{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: 0,
				Model:   req.Model,
				Choices: []StreamChoice{{
					Index: 0,
					Delta: ChatDelta{Content: delta},
					FinishReason: finishReason,
				}},
			}
		}
		if err := stream.Err(); err != nil {
			errChan <- err
		}
	}()

	return chunkChan, errChan
}
