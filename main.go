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
		Use:   "gencli",
		Short: "Gemini File CLI",
		Long: `
-----------------------------------------------------------------------------------------------------------------
Welcome to the Gemini-based interactive CLI tool!

This versatile command-line interface is designed to enhance your file management experience with :
		
	- Chat   : Engage in intelligent conversations with your ternimal.
	- Index  : Efficiently index your files for faster search and retrieval.
	- Search : Perform detailed searches to quickly find the files you need.
-----------------------------------------------------------------------------------------------------------------`,
		Version: version,
		Run: func(cmd *cobra.Command, arg []string) {
			fmt.Println(fileinfo.Green("\nWelcome to GenCLI\n"))
		},
	}

	// These functions need to written :
	rootCmd.AddCommand(cli.NewConfigCommand())
	rootCmd.AddCommand(cli.NewIndexCommand(hashSet))
	rootCmd.AddCommand(cli.NewSearchCommand())
	rootCmd.AddCommand(cli.NewChatCommand())

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(fileinfo.Red(err.Error()))
		return 1
	}

	// Save the hash set to file at the end
	if err := hashSet.SaveToFile(); err != nil {
		fmt.Println(fileinfo.Red(fmt.Sprintf("Error saving hash set: %v\n\nPlease Configurate\n", err)))
	}

	return 0

}

func main() {
	//start the application
	os.Exit(run())
}
