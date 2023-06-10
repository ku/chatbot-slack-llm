package messagestore

import "testing"

func TestSlackMessage_MentionAt(t *testing.T) {
	m := &SlackMessage{
		Text: "<!subteam^S01J9JZQZ8M> hello",
	}
	if !m.IsMentionAt("S01J9JZQZ8M") {
		t.Fatal()
	}
}
