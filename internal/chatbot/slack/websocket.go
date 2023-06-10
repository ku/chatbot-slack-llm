package slack

import (
	"context"
	"github.com/ku/chatbot/internal/chatbot"
	"github.com/ku/chatbot/messagestore"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
)

type websocket struct {
	listener chatbot.EventListener
	client   *slack.Client
}

var _ chatbot.ChatService = (*websocket)(nil)

type Slack interface {
	Run(botToken, appToken string) error
	PostMessageContext(ctx context.Context, m messagestore.Message, options ...slack.MsgOption) (string, string, error)
}

type WebsocketConfig struct{}

func NewWebsocket(conf *WebsocketConfig, client *slack.Client) *websocket {
	return &websocket{
		client: client,
	}
}

func (s *websocket) Name() string {
	return "slack.websocket"
}

func (w *websocket) SetEventListener(listener chatbot.EventListener) {
	w.listener = listener
}
func (s *websocket) Run(ctx context.Context) error {
	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		s.client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				// We have a new Events, let's type switch the event
				// Add more use cases here if you want to listen to other events.
				switch event.Type {
				// handle EventAPI events
				case socketmode.EventTypeEventsAPI:
					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// We need to send an Acknowledge to the slack server
					socketClient.Ack(*event.Request)
					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
					_, err := HandleEventMessage(ctx, s.listener, eventsAPIEvent)
					if err != nil {
						// Replace with actual err handeling
						log.Fatal(err)
					}
				case socketmode.EventTypeInteractive:
					cb, ok := event.Data.(slack.InteractionCallback)
					if !ok {
						log.Printf("Could not type cast the event to the InteractionCallback: %v\n", event)
						continue
					}
					err := s.listener.OnInteractionCallback(ctx, &cb)
					if err != nil {
						// Replace with actual err handeling
						log.Println(err)
					}
					socketClient.Ack(*event.Request)
				default:
					log.Printf("other event: %s", event.Type)
				}
			}
		}
	}(ctx, s.client, socketClient)

	return socketClient.Run()
}

func (s *websocket) PostReplyMessage(ctx context.Context, nm messagestore.Message) error {
	return postReplyMessage(ctx, s.client, nm)
}
