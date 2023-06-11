package slack

import (
	"context"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
)

func postMessageContext(ctx context.Context, client *slack.Client, m messagestore.Message, options ...slack.MsgOption) (string, string, error) {
	opts := []slack.MsgOption{
		slack.MsgOptionTS(m.GetThreadID()),
		slack.MsgOptionText(m.GetText(), false),
	}
	opts = append(opts, options...)
	return client.PostMessageContext(ctx, m.GetChannel(), opts...)
}
