package firestore

import (
	"time"

	"github.com/ku/chatbot-slack-llm/messagestore"
)

type conversation struct {
	Initiater string
	Channel   string
	ThreadID  string
	Messages  []firestoreMessage
}

type firestoreMessage struct {
	MessageID        string
	UserID           string
	Text             string
	MessageTimestamp string
	ThreadTimestamp  string
}

type message struct {
	conversation *conversation
	msg          *firestoreMessage
}

var _ messagestore.Message = (*message)(nil)

func (m *message) GetMessageID() string {
	return m.msg.MessageID
}

func (m *message) GetThreadID() string {
	return m.conversation.ThreadID
}

func (m *message) GetFrom() string {
	return m.msg.UserID
}

func (m *message) GetText() string {
	return m.msg.Text
}

func (m *message) GetRawText() string {
	return m.msg.Text
}

func (m *message) GetThreadTimestamp() string {
	return m.msg.ThreadTimestamp
}

func (m *message) GetTimestamp() string {
	return m.msg.MessageTimestamp
}

func (m *message) GetChannel() string {
	return m.conversation.Channel
}

func (m *message) GetCreatedAt() time.Time {
	return time.Now()
}

func (m *message) IsMentionAt(id string) bool {
	return messagestore.MentionAt(m.msg.Text) == id
}

func (c *conversation) GetMessages() []messagestore.Message {
	msgs := make([]messagestore.Message, len(c.Messages))
	for i, m := range c.Messages {
		msgs[i] = &message{
			conversation: c,
			msg:          &m,
		}
	}
	return msgs
}

func (c *conversation) IsFromInitiater(m messagestore.Message) bool {
	return c.Initiater == m.GetFrom()
}

func (c *conversation) String() string {
	return "firestore"
}
