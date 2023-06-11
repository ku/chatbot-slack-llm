package main

import (
	gospanner "cloud.google.com/go/spanner"
	"context"
	"fmt"
	"github.com/ku/chatbot/internal/chatbot"
	chatslack "github.com/ku/chatbot/internal/chatbot/slack"
	"github.com/ku/chatbot/internal/conversation/memory"
	"github.com/ku/chatbot/internal/conversation/spanner"
	"github.com/ku/chatbot/internal/llm"
	"github.com/ku/chatbot/internal/llm/openai"
	"github.com/ku/chatbot/internal/responder"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"os"
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

var opts struct {
	llm   string
	store string
	chat  string
}

func buildCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "chatbot",
		Short: "llm chatbot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return start()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&opts.llm, "llm", "l", "echo", "llm service [openai|echo]")
	rootCmd.PersistentFlags().StringVarP(&opts.store, "messagestore", "m", "memory", "messagestore [memory|spanner]")
	rootCmd.PersistentFlags().StringVarP(&opts.chat, "chat", "c", "websocket", "chat service [websocket|webhook]")
	return rootCmd
}

func start() error {
	ctx := context.Background()
	botID := os.Getenv("CHATBOT_BOT_ID")

	var llmClient chatbot.LLMClient
	var ms messagestore.MessageStore
	var chat chatbot.ChatService

	if opts.llm == "openai" {
		openaiApiKey := os.Getenv("OPENAI_API_KEY")
		llmClient = openai.NewClient(openaiApiKey)
	} else {
		llmClient = llm.NewEcho()
	}

	if opts.store == "spanner" {
		dsn := os.Getenv("CHATBOT_SPANNER_DSN")
		spc, err := gospanner.NewClient(ctx, dsn)
		if err != nil {
			return fmt.Errorf("failed to create spanner client: %w", err)
		}
		ms = spanner.NewConversations(botID, spc)
	} else {
		ms = memory.NewConversations(botID)
	}

	{
		botToken := os.Getenv("SLACK_BOT_TOKEN")
		appToken := os.Getenv("SLACK_APP_TOKEN")
		slackClient := slack.New(botToken, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

		if opts.chat == "websocket" {
			chat = chatslack.NewWebsocket(&chatslack.WebsocketConfig{}, slackClient)
		} else {
			chat = chatslack.NewWebHook(&chatslack.WebHookConfig{
				Addr:                  "localhost:3000",
				EventSubscriptionPath: "/subscription",
				InteractionPath:       "/interaction",
				SigningSecret:         os.Getenv("SLACK_SIGNING_SECRET"),
			}, slackClient)
		}
	}

	responder := responder.NewBashResponder()

	cb = chatbot.New(ms, chat, llmClient, responder, botID)
	chat.SetEventListener(cb)
	return chat.Run(ctx)
}
