package spanner

import (
	"cloud.google.com/go/spanner"
	"context"
	"github.com/ku/chatbot/internal/domains"
	"github.com/ku/chatbot/messagestore"
	"google.golang.org/grpc/codes"
	"math/rand"
	"strings"
)

type conversations struct {
	client *spanner.Client
}

var _ messagestore.MessageStore = (*conversations)(nil)

func NewConversations(client *spanner.Client) *conversations {
	return &conversations{client: client}
}

func (c *conversations) Name() string {
	return "spanner"
}

func (c *conversations) OnMention(ctx context.Context, m messagestore.Message) error {
	return c.OnMessage(ctx, m)
}

func (c *conversations) OnMessage(ctx context.Context, m messagestore.Message) error {
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
	if spanner.ErrCode(err) != codes.AlreadyExists {
		return err
	}
	return nil
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
