package proxy

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIAdapter struct {
	BaseURL string
}

func (a *OpenAIAdapter) SendChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (*ChatCompletionResponse, error) {
	config := openai.DefaultConfig(apiKey)
	if a.BaseURL != "" {
		config.BaseURL = a.BaseURL
	}

	client := openai.NewClientWithConfig(config)

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	})

	if err != nil {
		return nil, fmt.Errorf("openai error: %v", err)
	}

	choices := make([]ChatCompletionChoice, len(resp.Choices))
	for i, c := range resp.Choices {
		choices[i] = ChatCompletionChoice{
			Index: c.Index,
			Message: ChatMessage{
				Role:    c.Message.Role,
				Content: c.Message.Content,
			},
			FinishReason: string(c.FinishReason),
		}
	}

	return &ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}
func (a *OpenAIAdapter) StreamChatCompletion(ctx context.Context, req *ChatCompletionRequest, apiKey string) (<-chan *StreamResponseChunk, <-chan error) {
	chunkChan := make(chan *StreamResponseChunk)
	errChan := make(chan error, 1)

	config := openai.DefaultConfig(apiKey)
	if a.BaseURL != "" {
		config.BaseURL = a.BaseURL
	}
	client := openai.NewClientWithConfig(config)

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	}

	go func() {
		defer close(chunkChan)
		stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:       req.Model,
			Messages:    messages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		})
		if err != nil {
			errChan <- err
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err.Error() == "EOF" {
					return
				}
				errChan <- err
				return
			}

			choices := make([]StreamChoice, len(resp.Choices))
			for i, c := range resp.Choices {
				choices[i] = StreamChoice{
					Index: c.Index,
					Delta: ChatDelta{
						Role:    c.Delta.Role,
						Content: c.Delta.Content,
					},
					FinishReason: string(c.FinishReason),
				}
			}

			chunkChan <- &StreamResponseChunk{
				ID:      resp.ID,
				Object:  resp.Object,
				Created: resp.Created,
				Model:   resp.Model,
				Choices: choices,
			}
		}
	}()

	return chunkChan, errChan
}
