package conversation_test

import (
	gospanner "cloud.google.com/go/spanner"
	"context"
	"fmt"
	"github.com/ku/chatbot-slack-llm/chatbot"
	"github.com/ku/chatbot-slack-llm/internal/conversation/memory"
	"github.com/ku/chatbot-slack-llm/internal/conversation/spanner"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack/slackevents"
	"os"
	"testing"
	"time"
)

func TestConversations_Implementations(t *testing.T) {
	ctx := context.Background()
	botID := "S01J9JZQZ8M"
	dsn := os.Getenv("CHATBOT_SPANNER_DSN")
	client, err := gospanner.NewClient(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create spanner client: %s", err.Error())
	}

	impls := map[string]messagestore.MessageStore{
		"memory":  memory.NewConversations(botID),
		"spanner": spanner.NewConversations(botID, client),
	}

	for name, impl := range impls {
		t.Run(name+"/mention starts new conversation", func(t *testing.T) {
			added, err := impl.OnMessage(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "<!subteam^S01J9JZQZ8M> hello",
				TimeStamp:       "1685788142.416859",
				ThreadTimeStamp: "",
				Channel:         "c1",
			}))
			if err != nil {
				t.Fatalf("failed to OnMention: %s", err.Error())
			}

			if !added {
				t.Fatal("should be added")
			}

			if c, _ := impl.GetConversation(ctx, "1685788142.416859"); c == nil {
				t.Fatal("OnMention/GetConversation failed")
			}
		})

		t.Run(name+"/following messages in the thread are kept in the same conversation", func(t *testing.T) {
			now := time.Now().Unix()
			ts := fmt.Sprintf("%d.%d", now, now)
			added, err := impl.OnMessage(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "<!subteam^S01J9JZQZ8M> hello",
				TimeStamp:       ts,
				ThreadTimeStamp: "",
				Channel:         "c1",
			}))
			if err != nil {
				t.Fatalf("OnMention failed: %s", err.Error())
			}

			if !added {
				t.Fatal("should be added")
			}
			added, err = impl.OnMessage(ctx, chatbot.NewMessageFromMessage(&slackevents.MessageEvent{
				User:            "human",
				Text:            "are you listening?",
				TimeStamp:       "1685790093.186349",
				ThreadTimeStamp: ts,
				Channel:         "c1",
			}))
			if err != nil {
				t.Fatalf("OnMessage failed: %s", err.Error())
			}
			if !added {
				t.Fatal("second message should be added")
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
