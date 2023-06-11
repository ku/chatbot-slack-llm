package domains

import (
	"cloud.google.com/go/spanner"
	"context"
	"github.com/ku/chatbot/messagestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"time"
)

func (c *Conversation) GetThreadID() string {
	return c.ThreadID
}

func (c *Conversation) GetFrom() string {
	return c.ParentUserID
}

func (c *Conversation) GetText() string {
	return messagestore.FilterSlackMention(c.Text)
}

func (c *Conversation) GetRawText() string {
	return c.Text
}

func (c *Conversation) GetThreadTimestamp() string {
	return c.ThreadTimestamp
}

func (c *Conversation) GetMessageID() string {
	return c.GetMessageID()
}

func (c *Conversation) GetTimestamp() string {
	return c.MessageTimestamp
}

func (c *Conversation) GetChannel() string {
	return c.Channel
}

func (c *Conversation) GetCreatedAt() time.Time {
	return c.CreatedAt
}

// FindConversationsByThreadTimestampCreatedAt retrieves multiple rows from 'Conversations' as a slice of Conversation.
//
// Generated from index 'ConversationsByThreadID'.
func FindConversationsByThreadTimestamp(ctx context.Context, db YORODB, threadTimestamp string) ([]*Conversation, error) {
	const sqlstr = "SELECT " +
		"ConversationID, ParentUserID, Text, MessageTimestamp, ThreadTimestamp, ThreadID, Channel, CreatedAt " +
		"FROM Conversations@{FORCE_INDEX=ConversationsByThreadID} " +
		"WHERE ThreadID = @param0 ORDER BY CreatedAt ASC"

	stmt := spanner.NewStatement(sqlstr)
	stmt.Params["param0"] = threadTimestamp

	decoder := newConversation_Decoder(ConversationColumns())

	// run query
	YOLog(ctx, sqlstr, threadTimestamp)
	iter := db.Query(ctx, stmt)
	defer iter.Stop()

	// load results
	res := []*Conversation{}
	for {
		row, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, newError("FindConversationsByThreadTimestampCreatedAt", "Conversations", err)
		}

		c, err := decoder(row)
		if err != nil {
			return nil, newErrorWithCode(codes.Internal, "FindConversationsByThreadTimestampCreatedAt", "Conversations", err)
		}

		res = append(res, c)
	}

	return res, nil
}
