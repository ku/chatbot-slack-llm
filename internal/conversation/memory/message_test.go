package memory

import (
	"github.com/ku/chatbot-slack-llm/messagestore"
	"reflect"
	"testing"
	"time"
)

func TestMessage_Text(t *testing.T) {
	tests := map[string]struct {
		text string
		want string
	}{
		"@mention": {
			text: "<@U01J9JZQZ8M> hello",
			want: "hello",
		},
		"subteam": {
			text: "<!subteam^S01J9JZQZ8M> hello",
			want: "hello",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := &messagestore.SlackMessage{
				Text: tt.text,
			}
			if got := m.GetText(); got != tt.want {
				t.Errorf("Text() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_message_CreatedAt(t *testing.T) {
	m := &messagestore.SlackMessage{
		TS: "1685790080.750000128",
	}
	want := time.Date(2023, 6, 3, 20, 1, 20, 750000128, time.Local)
	if got := m.GetCreatedAt(); !reflect.DeepEqual(got, want) {
		t.Errorf("CreatedAt() = %v, want %v", got, want)
	}
}
