package cli

import (
	"context"
	"fmt"

	// "gemini_cli_tool/cli"
	"gemini_cli_tool/gemini"
	"os/user"

	"github.com/spf13/cobra"
)

type ChatOpts struct {
	Format     bool
	Style      string
	Multiline  bool
	Terminator string
}

var opts ChatOpts

func NewChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Starts a new chat session",
		RunE:  startChat,
	}

	cmd.Flags().BoolVarP(&opts.Format, "format", "f", true, "render markdown-formatted response")
	cmd.Flags().StringVarP(&opts.Style, "style", "s", "auto", "markdown format style (ascii, dark, light, pink, notty, dracula)")
	cmd.Flags().BoolVarP(&opts.Multiline, "multiline", "m", false, "read input as a multi-line string")
	cmd.Flags().StringVarP(&opts.Terminator, "term", "t", "~", "multi-line input terminator")

	return cmd
}

func startChat(cmd *cobra.Command, arg []string) error {
	fmt.Println("\nInitiating a new Chat Session..")

	// var opts ChatOpts
	// if err := cmd.Flags().Parse(&opts); err != nil {
	// 	return fmt.Errorf("failed to parse flags: %w", err)
	// }

	// fmt.Printf("Chat options: %+v\n", opts)

	// fmt.Println("Initializing chat session with Gemini API...")
	chatSession, err := gemini.NewChatSession(context.Background())
	if err != nil {
		// return err
		return fmt.Errorf("failed to initialize chat session: %w", err)
	}

	chat, err := NewChat(getCurrentUser(), chatSession, &opts)
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
