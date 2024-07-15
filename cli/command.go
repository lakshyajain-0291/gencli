package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var opts setupOpts

func NewConfigCommand() *cobra.Command {
	var addDirectories []string
	var deleteDirectories []string
	var addSkipTypes []string
	var deleteSkipTypes []string
	var addSkipFiles []string
	var deleteSkipFiles []string
	var relevanceIndex float32
	var showConfig bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Set configurations for the indexing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if showConfig {
				config, err := LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				showConfigFormatted(config)
				return nil
			}
			return setConfig(addDirectories, deleteDirectories, addSkipTypes, deleteSkipTypes, addSkipFiles, deleteSkipFiles, relevanceIndex)
		},
	}

	cmd.Flags().StringSliceVar(&addDirectories, "add-dir", []string{}, "List of directories to add to the index")
	cmd.Flags().StringSliceVar(&deleteDirectories, "del-dir", []string{}, "List of directories to remove from the index")
	cmd.Flags().StringSliceVar(&addSkipTypes, "add-skiptypes", []string{}, "List of file types to skip during indexing")
	cmd.Flags().StringSliceVar(&deleteSkipTypes, "del-skiptypes", []string{}, "List of file types to stop skipping during indexing")
	cmd.Flags().StringSliceVar(&addSkipFiles, "add-skipfiles", []string{}, "List of files to skip during indexing")
	cmd.Flags().StringSliceVar(&deleteSkipFiles, "del-skipfiles", []string{}, "List of files to stop skipping during indexing")
	cmd.Flags().Float32VarP(&relevanceIndex, "relindex", "r", 0.8, "Relevance Value used during Indexing")
	cmd.Flags().BoolVarP(&showConfig, "show-config", "s", false, "Show the current configuration")

	return cmd
}

func NewIndexCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index files in the configured directories",
		RunE:  indexFilesCmd,
	}

	return cmd
}

func NewSearchCommand() *cobra.Command {

	var allFileDisplay bool

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search files based on the provided query.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if allFileDisplay {
				return displayAllFiles()
			} else {
				return searchFilesCmd(cmd, args)
			}
		},
	}

	cmd.Flags().BoolVarP(&allFileDisplay, "all", "a", false, "Display Name and Description of All Indexed files")

	return cmd
}

func NewSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Starts a new setup session",
		RunE:  startsetup,
	}

	cmd.Flags().BoolVarP(&opts.Format, "format", "f", true, "render markdown-formatted response")
	cmd.Flags().StringVarP(&opts.Style, "style", "s", "auto", "markdown format style (ascii, dark, light, pink, notty, dracula)")
	cmd.Flags().BoolVarP(&opts.Multiline, "multiline", "m", false, "read input as a multi-line string")
	cmd.Flags().StringVarP(&opts.Terminator, "term", "t", "~", "multi-line input terminator")

	return cmd
}
