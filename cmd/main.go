package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	gospanner "cloud.google.com/go/spanner"
	"github.com/kelseyhightower/envconfig"
	"github.com/ku/chatbot-slack-llm/chatbot"
	slack2 "github.com/ku/chatbot-slack-llm/chatbot/slack"
	"github.com/ku/chatbot-slack-llm/internal/conversation/firestore"
	"github.com/ku/chatbot-slack-llm/internal/conversation/memory"
	"github.com/ku/chatbot-slack-llm/internal/conversation/spanner"
	"github.com/ku/chatbot-slack-llm/internal/llm"
	"github.com/ku/chatbot-slack-llm/internal/responder"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
)

type slackClientWrapper struct {
	client *slack.Client
}

var cb *chatbot.ChatBot

func main() {
	err := _main()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func _main() error {
	cmd := buildCommand()
	return cmd.Execute()
}

type Options struct {
	llm      string
	store    string
	protocol string
	webhook  string
	prompt   string
}

func buildCommand() *cobra.Command {
	var opts Options

	rootCmd := &cobra.Command{
		Use:   "chatbot",
		Short: "llm chatbot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(&opts)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&opts.llm, "llm", "l", "echo", "llm service [openai|echo]")
	rootCmd.PersistentFlags().StringVarP(&opts.store, "messagestore", "m", "memory", "messagestore [memory|spanner|firestore]")
	rootCmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "webhook", "protocol to receive events [websocket|webhook]")
	rootCmd.PersistentFlags().StringVarP(&opts.webhook, "webhook", "w", "", "use incoming webhook to send message")
	rootCmd.PersistentFlags().StringVarP(&opts.prompt, "prompt", "", "", `system prompt. text file name or URL of Slack File`)
	return rootCmd
}

func start(opts *Options) error {
	ctx := context.Background()
	botID := os.Getenv("CHATBOT_BOT_ID")

	var llmClient chatbot.LLMClient
	var ms messagestore.MessageStore
	var chat chatbot.ChatService
	var slackClient *slack.Client

	switch opts.store {
	case "spanner":
		dsn := os.Getenv("CHATBOT_SPANNER_DSN")
		spc, err := gospanner.NewClient(ctx, dsn)
		if err != nil {
			return fmt.Errorf("failed to create spanner client: %w", err)
		}
		ms = spanner.NewConversations(botID, spc)
	case "firestore":
		var conf firestore.FirestoreConfig
		conf.SlackBotID = botID
		err := envconfig.Process("", &conf)
		if err != nil {
			return fmt.Errorf("failed to init client: %w", err)
		}
		ms, err = firestore.New(ctx, &conf)
		if err != nil {
			return fmt.Errorf("failed to init firestore: %w", err)
		}
	default:
		ms = memory.NewConversations(botID)
	}

	var conf slack2.SlackConfig

	err := envconfig.Process("", &conf)
	if err != nil {
		return fmt.Errorf("failed to init client: %w", err)
	}

	{
		slackOpts := []slack.Option{
			slack.OptionDebug(true),
		}
		if opts.webhook != "" {
			slackOpts = append(slackOpts, slack.OptionAppLevelToken(conf.AppToken))
		}

		slackClient = slack.New(conf.BotToken, slackOpts...)

		if opts.protocol == "websocket" {
			chat = slack2.NewWebsocket(&slack2.WebsocketConfig{}, slackClient)
		} else {
			chat = slack2.NewWebHook(&slack2.WebHookConfig{
				SigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
				HTTP: &slack2.WebHookHTTPConfig{
					Addr:                  "localhost:3000",
					EventSubscriptionPath: "/subscription",
					InteractionPath:       "/interaction",
				},
			}, slackClient)
		}
	}

	promptURL, err := url.Parse(opts.prompt)
	//			opts.prompt
	if err != nil {
		return fmt.Errorf("failed to parse prompt value: %w", err)
	}
	if !(promptURL.Scheme == "" || promptURL.Scheme == "https") {
		return fmt.Errorf("invalid prompt value: %s", opts.prompt)
	}

	if opts.llm == "openai" {
		openaiApiKey := os.Getenv("OPENAI_API_KEY")
		llmClient = chatbot.NewOpenAIClient(openaiApiKey, func(ctx context.Context) (string, error) {
			if promptURL.Scheme == "" {
				b, err := os.ReadFile(opts.prompt)
				return string(b), err
			}

			return slack2.ReadCanvasAsPlainText(ctx, slackClient, promptURL)
		})
	} else {
		llmClient = llm.NewEcho()
	}

	responder := responder.NewBashResponder()

	cb = chatbot.New(ms, chat, llmClient, responder, botID, conf.SlackChannels)
	chat.SetEventListener(cb)
	return chat.Run(ctx)
}
