package cli

import (
	"context"
	"errors"
	"fmt"
	"os/user"
	"strings"

	"gemini_cli_tool/gemini"

	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

type chatOpts struct {
	Format     bool
	Style      string
	Multiline  bool
	Terminator string
}

type chat struct {
	session *gemini.Session
	prompt  *prompt
	reader  *readline.Instance
	opts    *chatOpts
}

func startchat(cmd *cobra.Command, arg []string) error {

	// Set the default API key
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	apiKeys := config.APIKeys
	if apiKeys == nil {
		return fmt.Errorf("no apikeys provided")
	}
	defaultApiKey := apiKeys[0]
	// fmt.Println("API : ", defaultApiKey)

	chatSession, err := gemini.NewchatSession(context.Background(), defaultApiKey)
	if err != nil {
		// return err
		return fmt.Errorf("failed to initialize chat session: %w", err)
	}

	chat, err := Newchat(getCurrentUser(), chatSession, &opts)
	if err != nil {
		// return err
		return fmt.Errorf("failed to create chat instance: %w", err)
	}

	// fmt.Println("Starting chat interface...")
	chat.Start()

	// fmt.Println("Closing chat session...")
	chatSession.Close()

	return nil
}

func getCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		return "user"
	}
	return currentUser.Username
}

func Newchat(user string, session *gemini.Session, opts *chatOpts) (*chat, error) {
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

	// fmt.Println("chat instance created.")
	return &chat{
		session: session,
		prompt:  prompt,
		reader:  reader,
		opts:    opts,
	}, nil
}

func (c *chat) Start() {
	// fmt.Println("chat started. Type 'exit' to quit.")
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

func (c *chat) read() (string, bool) {
	if c.opts.Multiline {
		return c.readMultiLine()
	}
	return c.readLine()
}

func (c *chat) readMultiLine() (string, bool) {
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

func (c *chat) readLine() (string, bool) {

	// fmt.Println("Reading single-line input...")
	input, err := c.reader.Readline()
	if err != nil {
		return c.handleReadError(input, err)
	}

	// fmt.Printf("Read line: %s\n", input)
	return validateInput(input)
}

func (c *chat) parseCommand(message string) command {

	// fmt.Printf("Parsing command: %s\n", message)
	if strings.HasPrefix(message, systemCmdPrefix) {
		return newSystemCommand(c)
	}
	return newGeminiCommand(c)
}

func (c *chat) handleReadError(input string, err error) (string, bool) {
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
