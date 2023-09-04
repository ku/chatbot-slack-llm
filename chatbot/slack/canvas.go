package slack

import (
	"bytes"
	"context"
	"fmt"
	"net/url"

	"github.com/slack-go/slack"
	"golang.org/x/net/html"
)

// TODO: get text from channelID.
// https://api.slack.com/methods/conversations.info/test
// https://api.slack.com/methods/files.info/test
// use private_url and bot token as authorizationn header of the request.
// https://stackoverflow.com/questions/36144761/access-slack-files-from-a-slack-bot
// curl -H 'authorization: Bearer xoxb-...' 'https://files.slack.com/files-pri/.../channel_canvas_'
func ReadCanvasAsPlainText(ctx context.Context, slackClient *slack.Client, canvasURL *url.URL) (string, error) {
	bw := new(bytes.Buffer)
	if err := slackClient.GetFileContext(ctx, canvasURL.String(), bw); err != nil {
		return "", err
	}

	doc, err := html.Parse(bw)
	if err != nil {
		return "", fmt.Errorf("failed to parse canvas as HTML document: %w", err)
	}

	var content string
	var f func(n *html.Node)
	f = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			content += n.Data
		case html.ElementNode:
			if n.Data == "br" {
				content += "\n"
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return content, nil
}

func extractText(n *html.Node, content *string) {
	if n.Type == html.TextNode {
		*content += n.Data
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, content)
	}
}
