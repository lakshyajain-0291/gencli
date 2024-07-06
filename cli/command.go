package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"google.golang.org/api/iterator"
)

const (
	systemCmdPrefix          = "$"
	systemCmdQuit            = "$q"
	systemCmdPurgeHistory    = "$p"
	systemCmdToggleInputMode = "$m"
	systemCmdToggleFormat    = "$tf"
	systemCmdSetStyle        = "$st"
)

type command interface {
	run(message string) bool
}

type systemCommand struct {
	chat *Chat
}

var _ command = (*systemCommand)(nil)

func newSystemCommand(chat *Chat) command {
	return &systemCommand{
		chat: chat,
	}
}

// run implements command.
func (c *systemCommand) run(message string) bool {
	switch message {
	case systemCmdQuit:
		c.print("Exiting gen-cli..")
		return true

	case systemCmdPurgeHistory:
		c.chat.session.ClearHistory()
		c.print("Cleared the chat history.")

	case systemCmdToggleInputMode:
		if c.chat.opts.Multiline {
			c.chat.reader.HistoryEnable()
			c.chat.opts.Multiline = false
			c.print("Switched to single-line input mode.")
		} else {
			c.chat.reader.HistoryDisable()
			c.chat.opts.Multiline = true
			c.print("Switched to Multi-line input mode.")

		}

	case systemCmdToggleFormat:
		c.chat.opts.Format = !c.chat.opts.Format
		c.print(fmt.Sprintf("Toggled format to %v.", c.chat.opts.Format))

	default:
		if strings.HasPrefix(message, systemCmdSetStyle) {
			style := strings.TrimPrefix(message, systemCmdSetStyle+" ")
			c.chat.opts.Style = style
			c.print(fmt.Sprintf("Set style to %s.", style))
		} else {
			c.print("Unknown system command.")
		}
	}

	return false
}

func (c *systemCommand) print(message string) {
	fmt.Printf("%s%s\n", c.chat.prompt.cli, message)
}

type geminiCommand struct {
	chat    *Chat
	spinner *spinner
	writer  *bufio.Writer
}

var _ command = (*geminiCommand)(nil)

func newGeminiCommand(chat *Chat) command {
	writer := bufio.NewWriter(os.Stdout)
	return &geminiCommand{
		chat:    chat,
		spinner: newSpinner(5, time.Second, writer),
		writer:  writer,
	}
}

// run implements command.
func (g *geminiCommand) run(message string) bool {
	g.printFlush(g.chat.prompt.gemini)
	g.spinner.start()

	if g.chat.opts.Format {
		g.runBlocking(message)
	} else {
		g.runStreaming(message)
	}

	return false
}

func (g *geminiCommand) runBlocking(message string) {
	// fmt.Println("Giving formatted output")

	response, err := g.chat.session.SendMessage(message)
	g.spinner.stop()

	if err != nil {
		fmt.Print(Red(err.Error()))
	} else {
		var builder strings.Builder

		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}

		output, err := glamour.Render(builder.String(), g.chat.opts.Style)
		if err != nil {
			fmt.Printf(Red("Failed to format : %s\n"), err)
			g.chat.opts.Format = false
			fmt.Println((builder.String()))
			return
		}

		fmt.Print(output)
	}
}

func (g *geminiCommand) runStreaming(message string) {
	// fmt.Println("Giving streaming output")

	responseIterator := g.chat.session.SendMessageStream(message)
	g.spinner.stop()
	for {
		response, err := responseIterator.Next()
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				fmt.Print(Red(err.Error()))
			}
			break
		}
		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				g.printFlush(fmt.Sprintf("%s", part))
			}
		}
	}

	fmt.Print("\n")
}

func (g *geminiCommand) printFlush(message string) {
	fmt.Fprintf(g.writer, "%s", message)
	g.writer.Flush()
}
