package cli

import (
	"errors"
	"fmt"
	"gemini_cli_tool/gemini"
	"strings"

	"github.com/chzyer/readline"
)

type Chat struct {
	session *gemini.ChatSession
	prompt  *prompt
	reader  *readline.Instance
	opts    *ChatOpts
}

func NewChat(user string, session *gemini.ChatSession, opts *ChatOpts) (*Chat, error) {
	// fmt.Println("Creating chat reader...")

	reader, err := readline.NewEx(&readline.Config{})
	if err != nil {
		return nil, err
	}

	// fmt.Println("Setting prompt...")
	prompt := newPrompt(user)

	reader.SetPrompt(prompt.user)

	if opts.Multiline {
		reader.HistoryDisable()
	}

	// fmt.Println("Chat instance created.")
	return &Chat{
		session: session,
		prompt:  prompt,
		reader:  reader,
		opts:    opts,
	}, nil
}

func (c *Chat) Start() {
	// fmt.Println("Chat started. Type 'exit' to quit.")
	for {
		message, ok := c.read()
		if !ok {
			continue
		}

		// fmt.Printf("Received message: %s\n", message)
		command := c.parseCommand(message)
		if quit := command.run(message); quit {
			break
		}
	}
}

func (c *Chat) read() (string, bool) {
	if c.opts.Multiline {
		return c.readMultiLine()
	}
	return c.readLine()
}

func (c *Chat) readMultiLine() (string, bool) {
	var builder strings.Builder

	term := c.opts.Terminator

	// fmt.Printf("Reading multi-line input, terminator: %s\n", term)
	for {
		input, err := c.reader.Readline()
		if err != nil {
			return c.handleReadError(input, err)
		}

		// fmt.Printf("Read line: %s\n", input)
		if strings.HasSuffix(input, term) {
			builder.WriteString(strings.TrimSuffix(input, term))
			break
		}
		if builder.Len() == 0 {
			c.reader.SetPrompt(c.prompt.userNext)
		}
		builder.WriteString(input + "\n")
	}

	c.reader.SetPrompt(c.prompt.user)
	return validateInput(builder.String())

}

func (c *Chat) readLine() (string, bool) {

	// fmt.Println("Reading single-line input...")
	input, err := c.reader.Readline()
	if err != nil {
		return c.handleReadError(input, err)
	}

	// fmt.Printf("Read line: %s\n", input)
	return validateInput(input)
}

func (c *Chat) parseCommand(message string) command {

	// fmt.Printf("Parsing command: %s\n", message)
	if strings.HasPrefix(message, systemCmdPrefix) {
		return newSystemCommand(c)
	}
	return newGeminiCommand(c)
}

func (c *Chat) handleReadError(input string, err error) (string, bool) {
	if errors.Is(err, readline.ErrInterrupt) {
		if len(input) == 0 {
			return systemCmdQuit, true
		}

		return "", false
	}
	fmt.Printf("%s%s\n", c.prompt.cli, err)
	return "", false
}

func validateInput(input string) (string, bool) {
	input = strings.TrimSpace(input)

	// fmt.Printf("Validated input: %s\n", input)
	return input, input != ""
}
