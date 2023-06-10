package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ku/chatbot/internal/chatbot"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io"
	"log"
	"net/http"
)

type webhook struct {
	conf     *WebHookConfig
	client   *slack.Client
	listener chatbot.EventListener
}

var _ chatbot.ChatService = (*webhook)(nil)

type WebHookConfig struct {
	Addr                  string
	EventSubscriptionPath string
	InteractionPath       string
	SigningSecret         string
}

func NewWebHook(conf *WebHookConfig, client *slack.Client) *webhook {
	return &webhook{
		conf:   conf,
		client: client,
	}
}

func (w *webhook) Name() string {
	return "slack.webhook"
}

func (w *webhook) SetEventListener(listener chatbot.EventListener) {
	w.listener = listener
}

func (w *webhook) Run(ctx context.Context) error {
	http.HandleFunc("/", wrap(func(_ http.ResponseWriter, req *http.Request) (any, error) {
		return nil, nil
	}))
	http.HandleFunc(w.conf.EventSubscriptionPath, wrap(func(_ http.ResponseWriter, req *http.Request) (any, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		if err := w.verifySignature(req, body); err != nil {
			return nil, err
		}

		ctx := req.Context()
		ctx = context.Background()

		return w.EventSubscriptionHandler(ctx, body)
	}))
	http.HandleFunc(w.conf.InteractionPath, func(w http.ResponseWriter, r *http.Request) {
	})
	fmt.Println("[INFO] Server listening")
	return http.ListenAndServe(w.conf.Addr, nil)
}

func (w *webhook) EventSubscriptionHandler(ctx context.Context, body []byte) (any, error) {
	eventsAPIEvent, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		return nil, err
	}

	return HandleEventMessage(ctx, w.listener, eventsAPIEvent)
}

func (w *webhook) InteractionHandler(ctx context.Context) error {
	return nil
}

func wrap(f func(w http.ResponseWriter, r *http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		switch v := resp.(type) {
		case string:
			w.Header().Set("content-type", "text")
			_, _ = w.Write([]byte(v))
		case []byte:
			w.Header().Set("content-type", "text")
			_, err = w.Write(v)
		default:
			w.Header().Set("content-type", "application/json")
			err = json.NewEncoder(w).Encode(resp)
		}
		if err != nil {
			log.Printf("%s %s: %s", r.Method, r.URL.Path, err.Error())
		}
	}
}

func (w *webhook) verifySignature(req *http.Request, body []byte) error {
	sv, err := slack.NewSecretsVerifier(req.Header, w.conf.SigningSecret)
	if err != nil {
		return err
	}
	if _, err := sv.Write(body); err != nil {
		return err
	}
	if err := sv.Ensure(); err != nil {
		return err
	}
	return nil
}

func (w *webhook) PostReplyMessage(ctx context.Context, message messagestore.Message) error {
	return postReplyMessage(ctx, w.client, message)
}
