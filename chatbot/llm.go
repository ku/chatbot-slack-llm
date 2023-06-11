package chatbot

import "github.com/ku/chatbot-slack-llm/internal/llm/openai"

func NewOpenAIClient(apiKey string, prompt func() (string, error)) *openai.Client {
	return openai.NewClient(apiKey, prompt)
}
