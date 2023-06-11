package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ku/chatbot/internal/chatbot"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log"
)

// HandleEventMessage will take an event and handle it properly based on the type of event
func HandleEventMessage(ctx context.Context, listener chatbot.EventListener, event slackevents.EventsAPIEvent) (any, error) {
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:
		var err error
		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// do nothing
		case *slackevents.MessageEvent:
			err = listener.OnMessage(ctx, ev)
		default:
			err = fmt.Errorf("unknown inner event type: %s", event.Type)
		}
		if err != nil {
			log.Println(err.Error())
		}
	case slackevents.URLVerification:
		vev, ok := event.Data.(*slackevents.EventsAPIURLVerificationEvent)
		if !ok {
			return nil, fmt.Errorf("could not type cast to ChallengeResponse: %v", event.Data)
		}
		println("verified")
		return vev.Challenge, nil
	default:
		return nil, fmt.Errorf("unsupported event type: %s", event.Type)
	}
	return nil, nil
}

func dumpAsJson(blocks []slack.Block) error {
	//test JSON on block kit builder
	//https://app.slack.com/block-kit-builder/
	msg := slack.NewBlockMessage(blocks...)
	b, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	log.Println(string(b))
	return nil
}

func postActionableMessage(ctx context.Context, slackClient *slack.Client, nm messagestore.Message) error {
	blocks, err := BuildBlocksFromResponse(nm)
	if err != nil {
		return err
	}
	o := slack.MsgOptionBlocks(blocks...)
	_, _, err = postMessageContext(ctx, slackClient, nm, o)
	return err
}
