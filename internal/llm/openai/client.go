package openai

import (
	"context"
	"fmt"

	"github.com/ku/chatbot-slack-llm/messagestore"
	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	client *openai.Client
	prompt func() (string, error)
}

type openaiCompletionResponse struct {
	resp *openai.ChatCompletionResponse
}

func (o *openaiCompletionResponse) GetText() string {
	if len(o.resp.Choices) == 0 {
		return ""
	}

	return o.resp.Choices[0].Message.Content
}

func NewClient(apiKey string, prompt func() (string, error)) *Client {
	return &Client{
		client: openai.NewClient(apiKey),
		prompt: prompt,
	}
}

func (c *Client) Name() string {
	return "openai"
}

func (c *Client) Completion(ctx context.Context, cv messagestore.Conversation) (messagestore.CompletionMessage, error) {
	var resp openai.ChatCompletionResponse
	var err error

	prompt, err := c.prompt()
	if err != nil {
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}
	msgs := conversationToMessages(cv, prompt)

	resp, err = c.client.CreateChatCompletion(
		ctx, openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    msgs,
			Temperature: 0,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	return &openaiCompletionResponse{resp: &resp}, nil
}

func conversationToMessages(cv messagestore.Conversation, prompt string) []openai.ChatCompletionMessage {
	msgs := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
	}

	for _, m := range cv.GetMessages() {
		var role string
		if cv.IsFromInitiater(m) {
			role = openai.ChatMessageRoleUser
		} else {
			role = openai.ChatMessageRoleAssistant
		}

		msgs = append(msgs, openai.ChatCompletionMessage{
			Role:    role,
			Content: m.GetText(),
		})
	}

	return msgs
}
