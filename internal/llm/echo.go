package llm

import (
	"context"

	"github.com/ku/chatbot-slack-llm/messagestore"
)

type echo struct{}

type echoMessage struct {
	msg string
}

func (e *echoMessage) GetText() string {
	return e.msg
}

func NewEcho() *echo {
	return &echo{}
}

func (e *echo) Name() string {
	return "echo"
}

func (e *echo) Completion(ctx context.Context, cv messagestore.Conversation) (messagestore.CompletionMessage, error) {
	msgs := cv.GetMessages()
	if len(msgs) == 0 {
		return &echoMessage{"howdy"}, nil
	}
	text := msgs[len(msgs)-1].GetText()
	return &echoMessage{text}, nil
}
