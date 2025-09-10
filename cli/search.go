package cli

import (
	"fmt"
	"gemini_cli_tool/fileinfo"
	"gemini_cli_tool/gemini"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func searchFilesCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no search query provided")
	}
	//   else {
	// 	fmt.Print(args[0])
	// }

	query := args[0]

	file, err := searchFiles(query)
	if err != nil {
		return err
	}

	fmt.Printf("\n%s \n\n%s %s\n\n%s %s\\%s\n\n%s %s\n", fileinfo.Green("Most relevelent file is -"), fileinfo.Yellow("File :"), file.Name, fileinfo.Yellow("File path :"), file.Directory, file.Name, fileinfo.Yellow("Description :"), file.Description)
	fmt.Print(fileinfo.Blue("\nEnter 'y' to open this file, or any other key to cancel: "))

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" {
		filepath := filepath.Join(file.Directory, file.Name)
		err = OpenFileWithDefaultApp(filepath)
		if err != nil {
			fmt.Print(fileinfo.Red(fmt.Sprintf("Failed to open file: %v", err)))
		}
	}

	return nil
}

func searchFiles(query string) (*fileinfo.FileInfo, error) {
	// spinners.start()

	files, err := LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index : %w", err)
	}

	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	apiKeys := config.APIKeys
	if apiKeys == nil {
		return nil, fmt.Errorf("no apikeys provided")
	}
	defaultApiKey := apiKeys[0]

	result, err := gemini.SearchRelevantFiles(files, query, config.RelevanceIndex, defaultApiKey)
	if err != nil {
		return nil, fmt.Errorf("search failed : %w", err)
	}

	// if len(result) == 0 {
	// 	fmt.Printf("No results found.")
	// } else {
	// 	fmt.Println("Most relevelnt files are : ")
	// 	for i := range results {
	// 		fmt.Printf("File : %s ,Directory : %s\n", files[results[i]].Name, files[results[i]].Description)
	// 	}
	// }

	if result < 0 {
		return nil, fmt.Errorf("no matches found")
	} else {
		return &files[result], nil
	}

}

func displayAllFiles() error {
	files, err := LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index : %w", err)
	}

	if files == nil {
		return fmt.Errorf("failed to find any files in index..ifileinfo.ndex the files")
	}

	for _, file := range files {
		fmt.Printf("\n%s %s\n\n%s %s\\%s\n\n%s %s\n", fileinfo.Yellow("File :"), file.Name, fileinfo.Yellow("File path :"), file.Directory, file.Name, fileinfo.Yellow("Description :"), file.Description)
		fmt.Println(fileinfo.Cyan("----------------------------------------------------------------------------------------------------------------------------------\n"))
	}

	return nil
}
