package memory

import (
	"context"
	"fmt"
	"github.com/ku/chatbot/messagestore"
)

type conversation struct {
	initiater string
	messages  []messagestore.Message
}

type conversations struct {
	botID string
	cvs   map[string]*conversation
}

var _ messagestore.Conversation = (*conversation)(nil)
var _ messagestore.MessageStore = (*conversations)(nil)

func NewConversations(botID string) *conversations {
	return &conversations{
		botID: botID,
		cvs:   make(map[string]*conversation),
	}
}

func (c *conversations) addNewConversation(m messagestore.Message) {
	_, ok := c.cvs[m.GetThreadID()]
	if !ok {
		c.cvs[m.GetThreadID()] = &conversation{
			messages: []messagestore.Message{m},
		}
	}
}

func (c *conversations) GetConversation(_ context.Context, thid string) (messagestore.Conversation, error) {
	cnvs, ok := c.cvs[thid]
	if !ok {
		return nil, fmt.Errorf("no conversation found for %s", thid)
	}
	return cnvs, nil
}

func (c *conversations) OnMention(ctx context.Context, m messagestore.Message) error {
	return nil
}

func (c *conversations) OnMessage(ctx context.Context, m messagestore.Message) error {
	cv, ok := c.cvs[m.GetThreadID()]
	if !ok {
		if !m.IsMentionAt(c.botID) {
			// received a random message. ignore it.
			return nil
		}
		cv = NewConversation(ctx, m)
		c.cvs[m.GetThreadID()] = cv
	}

	cv.AddMessage(ctx, m)
	return nil
}

func NewConversation(_ context.Context, m messagestore.Message) *conversation {
	return &conversation{
		initiater: m.GetFrom(),
		messages:  []messagestore.Message{m},
	}
}

func (c *conversations) Name() string {
	return "memory"
}

func (c *conversation) IsFromInitiater(m messagestore.Message) bool {
	return c.initiater == m.GetFrom()
}

func (c *conversation) AddMessage(_ context.Context, nm messagestore.Message) (bool, error) {
	for _, m := range c.messages {
		if m.GetTimestamp() == nm.GetTimestamp() {
			// already added.
			return false, nil
		}
	}

	c.messages = append(c.messages, nm)
	return true, nil
}

func (c *conversation) GetMessages() []messagestore.Message {
	return c.messages
}

func (c *conversation) String() string {
	var s string
	for _, m := range c.messages {
		s += m.GetRawText() + "\n"
	}
	return s
}
