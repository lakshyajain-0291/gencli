package main

import (
	"fmt"
	"gemini_cli_tool/cli"
	"gemini_cli_tool/fileinfo"
	"os"

	"github.com/spf13/cobra"
)

const (
	version   = "0.1.0"
	apiKeyEnv = "GEMINI_API_KEY"
)

func run() int {
	// fmt.Println("run..")

	hashSet := fileinfo.NewHashSet() // Initialize the hash set
	// Load the hash set from file at the start
	if err := hashSet.LoadFromFile(); err != nil {
		return 1
	}

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
	rootCmd.AddCommand(cli.NewConfigCommand())
	rootCmd.AddCommand(cli.NewIndexCommand(hashSet))
	rootCmd.AddCommand(cli.NewSearchCommand())
	rootCmd.AddCommand(cli.NewSetupCommand())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		return 1
	}

	// Save the hash set to file at the end
	if err := hashSet.SaveToFile(); err != nil {
		fmt.Printf("Error saving hash set: %v\n", err)
	}

	return 0

}

func main() {
	//start the application
	os.Exit(run())
}
