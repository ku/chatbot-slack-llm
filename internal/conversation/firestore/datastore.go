package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"github.com/ku/chatbot-slack-llm/messagestore"
)

var _ messagestore.Conversation = (*conversation)(nil)
var _ messagestore.MessageStore = (*Firestore)(nil)

type FirestoreConfig struct {
	Credential     string `envconfig:"FIRESTORE_CREDENTIAL"`
	ProjectID      string `envconfig:"FIRESTORE_PROJECT_ID"`
	CollectionName string `envconfig:"FIRESTORE_COLLECTION_NAME"`
	SlackBotID     string `envconfig:"SLACKBOT_ID"`
}

type Firestore struct {
	conf   *FirestoreConfig
	client *firestore.Client
	col    *firestore.CollectionRef
}

func New(ctx context.Context, conf *FirestoreConfig) (*Firestore, error) {
	opt := option.WithCredentialsJSON([]byte(conf.Credential))
	client, err := firestore.NewClient(ctx, conf.ProjectID, opt)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Firestore client: %w", err)
	}

	return &Firestore{
		client: client,
		conf:   conf,
		col:    client.Collection(conf.CollectionName),
	}, nil
}

func (f *Firestore) Name() string {
	return "Firestore"
}

func (f *Firestore) OnMessage(ctx context.Context, m messagestore.Message) (bool, error) {
	c := map[string]interface{}{
		"ThreadID": m.GetThreadID(),
		"Channel":  m.GetChannel(),
		"Messages": firestore.ArrayUnion(firestoreMessage{
			MessageID:        m.GetMessageID(),
			UserID:           m.GetFrom(),
			Text:             m.GetText(),
			MessageTimestamp: m.GetTimestamp(),
			ThreadTimestamp:  m.GetThreadTimestamp(),
		}),
	}
	if m.GetThreadTimestamp() == "" {
		c["Initiater"] = m.GetFrom()
	}

	_, err := f.col.Doc(m.GetThreadID()).Set(ctx, c,
		firestore.MergeAll,
	)

	return err == nil, err
}

func (f *Firestore) GetConversation(ctx context.Context, thid string) (messagestore.Conversation, error) {
	ss, err := f.col.Doc(thid).Get(ctx)
	if err != nil {
		return nil, err
	}
	var c conversation
	return &c, ss.DataTo(&c)
}
