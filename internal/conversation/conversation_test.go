package conversation_test

import (
	gospanner "cloud.google.com/go/spanner"
	"context"
	"fmt"
	"github.com/ku/chatbot/internal/chatbot"
	"github.com/ku/chatbot/internal/conversation/memory"
	"github.com/ku/chatbot/internal/conversation/spanner"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack/slackevents"
	"os"
	"testing"
	"time"
)

func TestConversations_Implementations(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("CHATBOT_SPANNER_DSN")
	client, err := gospanner.NewClient(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create spanner client: %s", err.Error())
	}

	impls := map[string]messagestore.MessageStore{
		"memory":  memory.NewConversations(),
		"spanner": spanner.NewConversations(client),
	}

	for name, impl := range impls {
		t.Run(name+"/mention starts new conversation", func(t *testing.T) {
			if err := impl.OnMention(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "hello",
				TimeStamp:       "1685788142.416859",
				ThreadTimeStamp: "",
				Channel:         "c1",
			})); err != nil {
				t.Fatalf("failed to OnMention: %s", err.Error())
			}

			if c, _ := impl.GetConversation(ctx, "1685788142.416859"); c == nil {
				t.Fatal("OnMention/GetConversation failed")
			}
		})

		t.Run(name+"/following messages in the thread are kept in the same conversation", func(t *testing.T) {
			now := time.Now().Unix()
			ts := fmt.Sprintf("%d.%d", now, now)
			if err := impl.OnMention(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "hello",
				TimeStamp:       ts,
				ThreadTimeStamp: "",
				Channel:         "c1",
			})); err != nil {
				t.Fatalf("OnMention failed: %s", err.Error())
			}

			if err := impl.OnMessage(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "are you listening?",
				TimeStamp:       "1685790093.186349",
				ThreadTimeStamp: ts,
				Channel:         "c1",
			})); err != nil {
				t.Fatalf("OnMessage failed: %s", err.Error())
			}

			cv, _ := impl.GetConversation(ctx, ts)
			if cv == nil {
				t.Fatal("conversation not found")
			}
			if msgs := cv.GetMessages(); len(msgs) != 2 {
				t.Fatalf("conversation has wrong number of messages. got %d", len(msgs))
			}
		})
	}

}
