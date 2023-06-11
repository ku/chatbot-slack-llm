package responder

import (
	"context"
	"fmt"
	"github.com/ku/chatbot/internal/chatbot"
	"io"
	"os/exec"
)

var _ chatbot.BlockActionResponder = (*BashResponder)(nil)

type BashResponder struct {
}

func NewBashResponder() *BashResponder {
	return &BashResponder{}
}

func (b *BashResponder) Handle(ctx context.Context, block string) (string, error) {
	cmd := exec.Command("/bin/bash")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	go func() {
		io.WriteString(stdin, block)
		stdin.Close()
	}()

	out, err := cmd.CombinedOutput()
	return string(out), err
}
