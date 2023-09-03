package spanner

import (
	"context"
	"math/rand"
	"strings"

	"cloud.google.com/go/spanner"

	"github.com/ku/chatbot-slack-llm/internal/domains"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"google.golang.org/grpc/codes"
)

type conversations struct {
	botID  string
	client *spanner.Client
}

var _ messagestore.MessageStore = (*conversations)(nil)

func NewConversations(botID string, client *spanner.Client) *conversations {
	return &conversations{
		botID:  botID,
		client: client,
	}
}

func (c *conversations) Name() string {
	return "spanner"
}

func (c *conversations) OnMessage(ctx context.Context, m messagestore.Message) (bool, error) {
	if m.GetThreadID() == "" {
		if !m.IsMentionAt(c.botID) {
			// received a random message. ignore it.
			return false, nil
		}
	}

	var added bool
	_, err := c.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		newrec := &domains.Conversation{
			ConversationID:   rand.Int63(),
			ParentUserID:     m.GetFrom(),
			Text:             m.GetRawText(),
			MessageTimestamp: m.GetTimestamp(),
			ThreadTimestamp:  m.GetThreadTimestamp(),
			ThreadID:         m.GetThreadID(),
			Channel:          m.GetChannel(),
			CreatedAt:        m.GetCreatedAt(),
		}

		m := newrec.Insert(ctx)
		err := txn.BufferWrite([]*spanner.Mutation{m})
		if err != nil {
			return nil
		}

		return nil
	})
	if err != nil {
		if spanner.ErrCode(err) != codes.AlreadyExists {
			return false, err
		}
	} else {
		added = true
	}
	return added, nil
}

func (c *conversations) GetConversation(ctx context.Context, conversationID string) (messagestore.Conversation, error) {
	ro := c.client.ReadOnlyTransaction()
	defer ro.Close()

	msgs, err := domains.FindConversationsByThreadTimestamp(ctx, ro, conversationID)
	if err != nil {
		return nil, err
	}

	return &conversation{
		msgs: msgs,
	}, nil
}

func (c *conversation) IsFromInitiater(m messagestore.Message) bool {
	if len(c.msgs) == 0 {
		return false
	}
	return c.msgs[0].ParentUserID == m.GetFrom()
}

type conversation struct {
	msgs []*domains.Conversation
}

func (c *conversation) GetMessages() []messagestore.Message {
	var msgs []messagestore.Message
	for _, m := range c.msgs {
		msgs = append(msgs, m)
	}
	return msgs
}

func (c *conversation) String() string {
	s := make([]string, len(c.msgs))
	for i, m := range c.msgs {
		s[i] = m.Text
	}
	return strings.Join(s, "\n")
}
