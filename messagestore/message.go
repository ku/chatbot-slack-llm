package messagestore

import (
	"context"
	"github.com/slack-go/slack/slackevents"
	"regexp"
	"time"
)

var mentionRegex = regexp.MustCompile(`<(!\w+\^|@)(\w+)>\s*`)

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

func NewMessageFromMention(ev *slackevents.AppMentionEvent) *SlackMessage {
	return &SlackMessage{
		RawMessage: ev,
		From:       ev.User,
		Text:       ev.Text,
		TS:         ev.TimeStamp,
		ThreadTS:   ev.TimeStamp,
		Channel:    ev.Channel,
		EventTS:    ev.EventTimeStamp,
	}
}

func NewMessageFromMessage(ev *slackevents.MessageEvent) *SlackMessage {
	return &SlackMessage{
		RawMessage: ev,
		From:       ev.User,
		Text:       ev.Text,
		TS:         ev.TimeStamp,
		ThreadTS:   ev.ThreadTimeStamp,
		Channel:    ev.Channel,
		EventTS:    ev.EventTimeStamp,
	}
}

func NewMessageFromCompletionMessage(channel string, thid string, m CompletionMessage) *SlackMessage {
	return &SlackMessage{
		From:     "",
		Text:     m.GetText(),
		ThreadTS: thid,
		Channel:  channel,
	}
}

func NewMessage(channel string, thid string, text string) *SlackMessage {
	return &SlackMessage{
		Channel:  channel,
		ThreadTS: thid,
		Text:     text,
	}
}
