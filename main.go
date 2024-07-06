package main

import (
	"fmt"
	"gemini_cli_tool/cli"
	"os"

	"github.com/spf13/cobra"
)

const (
	version   = "0.1.0"
	apiKeyEnv = "GEMINI_API_KEY"
)

func run() int {
	// fmt.Println("run..")

	rootCmd := &cobra.Command{
		Use:     "gencli",
		Short:   "Gemini File CLI",
		Long:    "Gemini Based interactable file manager CLI ",
		Version: version,
		Run: func(cmd *cobra.Command, arg []string) {
			fmt.Println("Welcome to gencli")
		},
	}

	// These functions need to written :
	// rootCmd.AddCommand(cmd.NewIndexCommand())
	// rootCmd.AddCommand(cmd.NewSearchCommand())
	rootCmd.AddCommand(cli.NewChatCommand())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	return 0

}

func main() {
	//start the application
	os.Exit(run())
}
