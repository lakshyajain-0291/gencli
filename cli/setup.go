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

type setupOpts struct {
	Format     bool
	Style      string
	Multiline  bool
	Terminator string
}

type setup struct {
	session *gemini.Session
	prompt  *prompt
	reader  *readline.Instance
	opts    *setupOpts
}

func startsetup(cmd *cobra.Command, arg []string) error {
	setupSession, err := gemini.NewsetupSession(context.Background())
	if err != nil {
		// return err
		return fmt.Errorf("failed to initialize setup session: %w", err)
	}

	setup, err := Newsetup(getCurrentUser(), setupSession, &opts)
	if err != nil {
		// return err
		return fmt.Errorf("failed to create setup instance: %w", err)
	}

	// fmt.Println("Starting setup interface...")
	setup.Start()

	// fmt.Println("Closing setup session...")
	setupSession.Close()

	return nil
}

func getCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		return "user"
	}
	return currentUser.Username
}

func Newsetup(user string, session *gemini.Session, opts *setupOpts) (*setup, error) {
	// fmt.Println("Creating setup reader...")

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

	// fmt.Println("setup instance created.")
	return &setup{
		session: session,
		prompt:  prompt,
		reader:  reader,
		opts:    opts,
	}, nil
}

func (c *setup) Start() {
	// fmt.Println("setup started. Type 'exit' to quit.")
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

func (c *setup) read() (string, bool) {
	if c.opts.Multiline {
		return c.readMultiLine()
	}
	return c.readLine()
}

func (c *setup) readMultiLine() (string, bool) {
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

func (c *setup) readLine() (string, bool) {

	// fmt.Println("Reading single-line input...")
	input, err := c.reader.Readline()
	if err != nil {
		return c.handleReadError(input, err)
	}

	// fmt.Printf("Read line: %s\n", input)
	return validateInput(input)
}

func (c *setup) parseCommand(message string) command {

	// fmt.Printf("Parsing command: %s\n", message)
	if strings.HasPrefix(message, systemCmdPrefix) {
		return newSystemCommand(c)
	}
	return newGeminiCommand(c)
}

func (c *setup) handleReadError(input string, err error) (string, bool) {
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
