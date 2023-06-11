package messagestore

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SlackMessage struct {
	RawMessage interface{}
	From       string
	Text       string
	TS         string
	ThreadTS   string
	Channel    string
	EventTS    string
}

var _ Message = (*SlackMessage)(nil)

func FilterSlackMention(s string) string {
	return mentionRegex.ReplaceAllString(s, "")
}

// Text() replaces mention meta texts with empty string.
// https://api.slack.com/reference/surfaces/formatting
func (m *SlackMessage) GetText() string {
	return FilterSlackMention(m.Text)
}

func (m *SlackMessage) GetFrom() string {
	return m.From
}

// GetMessageID returns unique id of the message.
func (m *SlackMessage) GetMessageID() string {
	return m.GetTimestamp()
}

func (m *SlackMessage) GetTimestamp() string {
	return m.TS
}

func (m *SlackMessage) GetRawText() string {
	return m.Text
}

func (m *SlackMessage) GetChannel() string {
	return m.Channel
}

func (m *SlackMessage) GetThreadID() string {
	if m.ThreadTS == "" {
		return m.TS
	}
	return m.ThreadTS
}

func (m *SlackMessage) String() string {
	return fmt.Sprintf("%s: %s", m.From, m.Text)
}

func (m *SlackMessage) GetThreadTimestamp() string {
	return m.ThreadTS
}

func (m *SlackMessage) GetCreatedAt() time.Time {
	a := strings.Split(m.TS, ".")
	if len(a) != 2 {
	}

	f, err := strconv.ParseFloat(m.TS, 64)
	if err != nil {
		return time.Time{}
	}

	return time.Unix(0, 0).Add(time.Duration(f * float64(time.Second)))
}

func (m *SlackMessage) IsMentionAt(id string) bool {
	return MentionAt(m.Text) == id
}

func MentionAt(s string) string {
	m := mentionRegex.FindStringSubmatch(s)
	if len(m) == 3 {
		return m[2]
	}
	return ""
}
