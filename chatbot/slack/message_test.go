package slack_test

import (
	"github.com/ku/chatbot/chatbot"
	"github.com/ku/chatbot/chatbot/slack"
	"testing"
)

func TestBuildBlocksFromResponse(t *testing.T) {
	s := "first user:\n\n```\nchatbot create \nchatbot add title\nchatbot add description\n```\n\nsecond user:\n\n```\nchatbot create\nchatbot add title\nchatbot add description\n```"
	m := chatbot.NewMessage("channel", "thread", s)
	got, err := slack.BuildBlocksFromResponse(m)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(got))
	}
}
