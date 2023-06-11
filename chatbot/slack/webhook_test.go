package slack

import (
	"bytes"
	"context"
	"github.com/ku/chatbot-slack-llm/chatbot"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type echoResponder struct{}

func (e echoResponder) Handle(ctx context.Context, block string) (string, error) {
	return block, nil
}

type chatmock struct {
	msg messagestore.Message
}

func (c *chatmock) Name() string { return "chatmock" }
func (c *chatmock) PostMessage(ctx context.Context, message messagestore.Message) error {
	c.msg = message
	return nil
}
func (c *chatmock) PostActionableMessage(ctx context.Context, message messagestore.Message) error {
	return nil
}
func (c *chatmock) SetEventListener(listener chatbot.EventListener) {}
func (c *chatmock) Run(ctx context.Context) error                   { return nil }

func Test_webhook_InteractivityHandler(t *testing.T) {

	tests := map[string]struct {
		payload string
		want    string
		wantErr bool
	}{
		"echo": {
			payload: "{\"type\":\"block_actions\",\"user\":{\"id\":\"U7FEW22AV\",\"username\":\"ku0522a\",\"name\":\"ku0522a\",\"team_id\":\"T7G0NU082\"},\"api_app_id\":\"A050M7A6L3T\",\"token\":\"8kNEn4ijXcD9lsazlHkUzIyV\",\"container\":{\"type\":\"message\",\"message_ts\":\"1686450056.622089\",\"channel_id\":\"C7Q885D4P\",\"is_ephemeral\":false,\"thread_ts\":\"1686450055.262239\"},\"trigger_id\":\"5391783133831.254022952274.50ef0f34fa6045a4a40ad14f1c284624\",\"team\":{\"id\":\"T7G0NU082\",\"domain\":\"arbitd\"},\"enterprise\":null,\"is_enterprise_install\":false,\"channel\":{\"id\":\"C7Q885D4P\",\"name\":\"alert\"},\"message\":{\"bot_id\":\"B05020UG8MU\",\"type\":\"message\",\"text\":\"``` echo \\\"hello\\\" ```\",\"user\":\"U050Y9GHYCQ\",\"ts\":\"1686450056.622089\",\"app_id\":\"A050M7A6L3T\",\"blocks\":[{\"type\":\"section\",\"block_id\":\"+Jqxa\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"```echo \\\"hello\\\"```\",\"verbatim\":false},\"accessory\":{\"type\":\"button\",\"action_id\":\"run-\",\"text\":{\"type\":\"plain_text\",\"text\":\"Run\",\"emoji\":true},\"value\":\"echo \\\"hello\\\"\"}}],\"team\":\"T7G0NU082\",\"thread_ts\":\"1686450055.262239\",\"parent_user_id\":\"U7FEW22AV\"},\"state\":{\"values\":{}},\"response_url\":\"https:\\/\\/hooks.slack.com\\/actions\\/T7G0NU082\\/5399664778902\\/67lk8f2RiwhSPkAOkXc0q2ao\",\"actions\":[{\"action_id\":\"run-\",\"block_id\":\"+Jqxa\",\"text\":{\"type\":\"plain_text\",\"text\":\"Run\",\"emoji\":true},\"value\":\"echo \\\"hello\\\"\",\"type\":\"button\",\"action_ts\":\"1686450088.658086\"}]}",
			want:    "```echo \"hello\"```\n```echo \"hello\"```",
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			chat := &chatmock{}
			bot := chatbot.New(nil, chat, nil, &echoResponder{}, "")
			webhook := &WebHook{
				listener: bot,
			}
			f := webhook.InteractivityHandler()

			val := url.Values{}
			val.Set("payload", tt.payload)

			req := (&http.Request{
				Body: io.NopCloser(bytes.NewReader([]byte(val.Encode()))),
			}).WithContext(ctx)

			rec := &httptest.ResponseRecorder{}
			f(rec, req)

			got := chat.msg.GetText()

			if rec.Code != 200 {
				t.Fatal("http request failed")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractivityHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
