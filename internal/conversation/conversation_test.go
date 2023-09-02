package conversation_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	gospanner "cloud.google.com/go/spanner"
	"github.com/kelseyhightower/envconfig"
	"github.com/ku/chatbot-slack-llm/internal/conversation/firestore"
	"github.com/ku/chatbot-slack-llm/internal/conversation/memory"
	"github.com/ku/chatbot-slack-llm/internal/conversation/spanner"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack/slackevents"
)

type initializer func() (messagestore.MessageStore, error)

func TestConversations_Implementations(t *testing.T) {
	ctx := context.Background()
	botID := "S01J9JZQZ8M"

	inits := map[string]initializer{
		"firestore": func() (messagestore.MessageStore, error) {
			ctx := context.Background()
			var conf firestore.FirestoreConfig
			conf.SlackBotID = botID
			if err := envconfig.Process("", &conf); err != nil {
				return nil, fmt.Errorf("failed to init client: %w", err)
			}
			return firestore.New(ctx, &conf)
		},
		"memory": func() (messagestore.MessageStore, error) {
			return memory.NewConversations(botID), nil
		},
		"spanner": func() (messagestore.MessageStore, error) {
			dsn := os.Getenv("CHATBOT_SPANNER_DSN")
			client, err := gospanner.NewClient(ctx, dsn)
			if err != nil {
				t.Fatalf("failed to create spanner client: %s", err.Error())
			}
			return spanner.NewConversations(botID, client), nil
		},
	}

	for name, init := range inits {
		impl, err := init()
		if err != nil {
			t.Errorf("failed to init %s: %s", name, err.Error())
		}

		t.Run(name+"/mention starts new conversation", func(t *testing.T) {
			added, err := impl.OnMessage(ctx, messagestore.NewMessageFromMessage(&slackevents.MessageEvent{
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
			added, err := impl.OnMessage(ctx, messagestore.NewMessageFromMessage(&slackevents.MessageEvent{
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
			added, err = impl.OnMessage(ctx, messagestore.NewMessageFromMessage(&slackevents.MessageEvent{
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
