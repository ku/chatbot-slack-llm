package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ku/chatbot-slack-llm/chatbot"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type WebHook struct {
	conf             *WebHookConfig
	client           *slack.Client
	listener         chatbot.EventListener
	llmTimeout       time.Duration
	responderTimeout time.Duration
}

var _ chatbot.ChatService = (*WebHook)(nil)

type WebHookConfig struct {
	SigningSecret string
	HTTP          *WebHookHTTPConfig
}

type WebHookHTTPConfig struct {
	Addr                  string
	EventSubscriptionPath string
	InteractionPath       string
}

func NewWebHook(conf *WebHookConfig, client *slack.Client) *WebHook {
	return &WebHook{
		conf:   conf,
		client: client,
	}
}

func (w *WebHook) Name() string {
	return "slack.webhook"
}

func (w *WebHook) SetEventListener(listener chatbot.EventListener) {
	w.listener = listener
}

func (w *WebHook) Run(ctx context.Context) error {
	http.HandleFunc("/", wrap(func(_ http.ResponseWriter, req *http.Request) (any, error) {
		return nil, nil
	}))
	http.HandleFunc(w.conf.HTTP.EventSubscriptionPath, w.EventSubscriptionHandler())
	http.HandleFunc(w.conf.HTTP.InteractionPath, w.InteractivityHandler())
	return http.ListenAndServe(w.conf.HTTP.Addr, nil)
}

func (w *WebHook) InteractivityHandler() func(http.ResponseWriter, *http.Request) {
	return wrap(func(_ http.ResponseWriter, req *http.Request) (any, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %w", err)
		}
		payload := vals.Get("payload")

		ctx := req.Context()
		// LLM takes a long time to respond, so use another context here.
		ctx = context.Background()

		return w.interactivityHandler(ctx, []byte(payload))
	})
}

func (w *WebHook) EventSubscriptionHandler() func(http.ResponseWriter, *http.Request) {
	return wrap(func(_ http.ResponseWriter, req *http.Request) (any, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		if err := w.verifySignature(req, body); err != nil {
			return nil, err
		}

		return w.eventSubscriptionHandler(req.Context(), body)
	})
}

func (w *WebHook) eventSubscriptionHandler(ctx context.Context, body []byte) (any, error) {
	eventsAPIEvent, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		return nil, err
	}

	return HandleEventMessage(ctx, w.listener, eventsAPIEvent)
}

// InteractivityHandler handles interactive events like pressing buttons.
func (w *WebHook) interactivityHandler(ctx context.Context, b []byte) (any, error) {
	var icb slack.InteractionCallback
	if err := json.Unmarshal(b, &icb); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	err := w.listener.OnInteractionCallback(ctx, &icb)
	if err != nil {
		// Replace with actual err handling
		log.Println(err)
	}
	return nil, err
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

func (w *WebHook) verifySignature(req *http.Request, body []byte) error {
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

func (w *WebHook) PostMessage(ctx context.Context, message messagestore.Message) error {
	_, _, err := postMessageContext(ctx, w.client, message)
	return err
}

func (w *WebHook) PostActionableMessage(ctx context.Context, message messagestore.Message) error {
	return postActionableMessage(ctx, w.client, message)
}
