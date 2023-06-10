package slack_test

import (
	"github.com/ku/chatbot/internal/chatbot/slack"
	"testing"
)

func TestBuildBlocksFromResponse(t *testing.T) {
	s := "first user:\n\n```\nchatbot create \nchatbot add title\nchatbot add description\n```\n\nsecond user:\n\n```\nchatbot create\nchatbot add title\nchatbot add description\n```"

	got, err := slack.BuildBlocksFromResponse(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatal("expected 5")
	}
}
