package chatbot

import (
	"context"
	"fmt"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log"
	"os"
	"os/exec"
	"strings"
)

type ChatBot struct {
	llm       LLMClient
	store     messagestore.MessageStore
	chat      ChatService
	responder BlockActionResponder

	botID   string
	verbose bool
}

type ChatService interface {
	Name() string
	PostMessage(ctx context.Context, message messagestore.Message) error
	PostActionableMessage(ctx context.Context, message messagestore.Message) error
	SetEventListener(listener EventListener)
	Run(ctx context.Context) error
}

type LLMClient interface {
	Name() string
	Completion(ctx context.Context, cv messagestore.Conversation, prompt string) (messagestore.CompletionMessage, error)
}

type EventListener interface {
	OnMessage(ctx context.Context, ev *slackevents.MessageEvent) error
	OnInteractionCallback(ctx context.Context, acbs *slack.InteractionCallback) error
}

type BlockActionResponder interface {
	Handle(ctx context.Context, block string) (string, error)
}

func New(store messagestore.MessageStore, chat ChatService, llm LLMClient, responder BlockActionResponder, botID string) *ChatBot {
	return &ChatBot{
		llm:       llm,
		store:     store,
		chat:      chat,
		responder: responder,
		botID:     botID,
	}
}

func (c *ChatBot) GetConversation(thid string) messagestore.Conversation {
	cnvs, err := c.store.GetConversation(context.Background(), thid)
	if err != nil {
		return nil
	}
	return cnvs
}

func prompt() string {
	b, err := os.ReadFile("./prompt.txt")
	if err != nil {
		panic(fmt.Sprintf("failed to read prompt.txt: %v", err))
	}
	return string(b)
}

func (c *ChatBot) OnMention(ctx context.Context, ev *slackevents.AppMentionEvent) error {
	return nil
}

func (c *ChatBot) OnMessage(ctx context.Context, ev *slackevents.MessageEvent) error {
	m := NewMessageFromMessage(ev)

	_ = c.processDebugMessage(ctx, m)

	// ignore messages from myself
	if c.botID == m.GetFrom() {
		return nil
	}

	added, err := c.store.OnMessage(ctx, m)
	if !added {
		return err
	}

	cv, err := c.store.GetConversation(ctx, m.GetThreadID())
	if err != nil {
		return err
	}

	if c.shouldIgnore(cv, m) {
		return nil
	}

	func() {
		if err := c.respondToMessage(ctx, cv, m); err != nil {
			log.Println(err.Error())
		}
	}()

	return nil
}

func (c *ChatBot) postReply(ctx context.Context, nm messagestore.Message) error {
	return c.chat.PostActionableMessage(ctx, nm)
}

func (c *ChatBot) OnInteractionCallback(ctx context.Context, cb *slack.InteractionCallback) error {
	for _, ba := range cb.ActionCallback.BlockActions {
		var exitStatus string
		script := ba.Value

		output, err := c.responder.Handle(ctx, script)
		// report the result
		if err != nil {
			ee, ok := err.(*exec.ExitError)
			if !ok {
				return err
			}
			exitStatus = fmt.Sprintf("%s\n", ee.Error())
		} else {
			exitStatus = ""
		}
		thid := cb.Message.Msg.ThreadTimestamp

		msg := fmt.Sprintf("```%s```\n%s```%s```", script, exitStatus, output)
		if err := c.chat.PostMessage(ctx, NewMessage(cb.Channel.ID, thid, msg)); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChatBot) processDebugMessage(ctx context.Context, m messagestore.Message) error {
	if !strings.HasPrefix(m.GetText(), "debug ") {
		return nil
	}

	if m.GetText() == "debug vars" {
		vars := []string{
			"mode: " + c.chat.Name(),
			"messagestorage: " + c.store.Name(),
			"llm: " + c.llm.Name(),
			"botID: " + c.botID,
		}

		if err := c.chat.PostActionableMessage(ctx, NewMessage(
			m.GetChannel(),
			m.GetThreadID(),
			"^DEBUG\n"+strings.Join(vars, "\n"),
		)); err != nil {
			return err
		}

		return nil
	}

	if m.GetText() == "debug off" {
		c.verbose = false
		return nil
	}

	if m.GetText() == "debug on" {
		c.verbose = true
	}

	if c.verbose {
		cv, err := c.store.GetConversation(ctx, m.GetThreadID())
		if err != nil {
			return err
		}

		s := cv.String()
		if err := c.chat.PostActionableMessage(ctx, NewMessage(
			m.GetChannel(),
			m.GetThreadID(),
			"^DEBUG\n"+s,
		)); err != nil {
			return err
		}
	}
	return nil
}

func (c *ChatBot) respondToMessage(ctx context.Context, cv messagestore.Conversation, m messagestore.Message) error {
	resp, err := c.llm.Completion(ctx, cv, prompt())
	if err != nil {
		return fmt.Errorf("llm completion failed: %w", err)
	}
	nm := NewMessageFromCompletionMessage(m.GetChannel(), m.GetThreadID(), resp)

	if err := c.postReply(ctx, nm); err != nil {
		return err
	}

	return nil
}

func (c *ChatBot) shouldIgnore(cv messagestore.Conversation, nm messagestore.Message) bool {
	msgs := cv.GetMessages()
	if len(msgs) == 0 {
		return true
	}

	if !cv.IsFromInitiater(nm) {
		return true
	}

	firstMsg := msgs[0]
	// if it's a mention and the first message
	if nm.IsMentionAt(c.botID) {
		if len(msgs) == 1 && firstMsg.GetTimestamp() == nm.GetTimestamp() {
			return false
		}
	} else {
		// not a mention but a message in a thread which started by a mention
		if firstMsg.IsMentionAt(c.botID) && firstMsg.GetFrom() == nm.GetFrom() {
			return false
		}
	}

	return true
}
