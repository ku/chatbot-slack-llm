package slack

import (
	"fmt"
	"github.com/ku/chatbot-slack-llm/messagestore"
	"github.com/slack-go/slack"
	"strings"
)

type ResponseBlockType int

const (
	ResponseBlockTypeText     ResponseBlockType = iota
	ResponseBlockTypeCommands ResponseBlockType = iota
)

type ResponseBlock struct {
	Type ResponseBlockType
	Text string
}

func BuildBlocksFromResponse(m messagestore.Message) ([]slack.Block, error) {
	blocks := []slack.Block{}
	mid := m.GetMessageID()
	s := m.GetText()
	responseBlocks := CommandBlocksFromResponse(s)
	for _, block := range responseBlocks {
		s := block.Text
		if block.Type == ResponseBlockTypeCommands {
			runBtnText := slack.NewTextBlockObject("plain_text", "Run", true, false)
			runBtnEle := slack.NewButtonBlockElement(fmt.Sprintf("run-%s", mid), block.Text, runBtnText)

			text := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("```%s```", block.Text), false, false)
			section := slack.NewSectionBlock(text, nil, slack.NewAccessory(runBtnEle))
			blocks = append(blocks, section)
			continue
		}

		text := slack.NewTextBlockObject("mrkdwn", s, false, false)
		section := slack.NewSectionBlock(text, nil, nil)
		blocks = append(blocks, section)
	}

	return blocks, nil
}

func CommandBlocksFromResponse(rawText string) []*ResponseBlock {
	var blocks []*ResponseBlock
	//Replace the ampersand, &, with &amp;
	//Replace the less-than sign, < with &lt;
	//Replace the greater-than sign, > with &gt;
	// https://api.slack.com/reference/surfaces/formatting#escaping
	unescapedTexet := strings.ReplaceAll(rawText, "&amp;", "&")
	fields := strings.Split(unescapedTexet, "```")
	for n, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if (n % 2) == 0 {
			blocks = append(blocks, &ResponseBlock{
				Type: ResponseBlockTypeText,
				Text: field,
			})
		} else {
			//  in  ``` block
			blocks = append(blocks, &ResponseBlock{
				Type: ResponseBlockTypeCommands,
				Text: field,
			})
		}
	}
	return blocks
}
