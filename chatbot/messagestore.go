package chatbot

import (
	"cloud.google.com/go/spanner"
	chatbotspanner "github.com/ku/chatbot-slack-llm/internal/conversation/spanner"
	"github.com/ku/chatbot-slack-llm/messagestore"
)

func NewSpannerMessageStore(client *spanner.Client, botID string) messagestore.MessageStore {
	return chatbotspanner.NewConversations(botID, client)
}
