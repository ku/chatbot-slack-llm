package messagestore

import (
	"context"
	"time"
)

type MessageStore interface {
	Name() string
	OnMessage(ctx context.Context, m Message) (bool, error)
	GetConversation(ctx context.Context, thid string) (Conversation, error)
}

type Conversation interface {
	GetMessages() []Message
	IsFromInitiater(m Message) bool
	String() string
}

type Message interface {
	GetMessageID() string
	GetThreadID() string
	GetFrom() string
	GetText() string
	GetRawText() string
	GetThreadTimestamp() string
	GetTimestamp() string
	GetChannel() string
	GetCreatedAt() time.Time
	IsMentionAt(id string) bool
}

type CompletionMessage interface {
	GetText() string
}
