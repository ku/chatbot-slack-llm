package chatbot

import (
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack/slackevents"
)

func NewMessageFromMention(ev *slackevents.AppMentionEvent) *messagestore.SlackMessage {
	return &messagestore.SlackMessage{
		RawMessage: ev,
		From:       ev.User,
		Text:       ev.Text,
		TS:         ev.TimeStamp,
		ThreadTS:   ev.TimeStamp,
		Channel:    ev.Channel,
		EventTS:    ev.EventTimeStamp,
	}
}

func NewMessageFromMessage(ev *slackevents.MessageEvent) *messagestore.SlackMessage {
	return &messagestore.SlackMessage{
		RawMessage: ev,
		From:       ev.User,
		Text:       ev.Text,
		TS:         ev.TimeStamp,
		ThreadTS:   ev.ThreadTimeStamp,
		Channel:    ev.Channel,
		EventTS:    ev.EventTimeStamp,
	}
}

func NewMessageFromCompletionMessage(channel string, thid string, m messagestore.CompletionMessage) *messagestore.SlackMessage {
	return &messagestore.SlackMessage{
		From:     "",
		Text:     m.GetText(),
		ThreadTS: thid,
		Channel:  channel,
	}
}

func NewMessage(channel string, thid string, text string) *messagestore.SlackMessage {
	return &messagestore.SlackMessage{
		Channel:  channel,
		ThreadTS: thid,
		Text:     text,
	}
}
