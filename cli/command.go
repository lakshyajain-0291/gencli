package cli

import (
	"github.com/spf13/cobra"
)

var opts setupOpts

func NewConfigCommand() *cobra.Command {
	var directories []string
	var skipTypes []string
	var skipFiles []string
	var relevanceIndex float32

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Set configurations for the indexing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setConfig(directories, skipTypes, skipFiles, relevanceIndex)
		},
	}

	cmd.Flags().StringSliceVarP(&directories, "directories", "d", []string{}, "List of directories to index")
	cmd.Flags().StringSliceVarP(&skipTypes, "skiptypes", "t", []string{}, "List of file types to skip during indexing")
	cmd.Flags().StringSliceVarP(&skipFiles, "skipfiles", "f", []string{}, "List of directories to index")
	cmd.Flags().Float32VarP(&relevanceIndex, "relativeindex", "r", 0.8, "Relevance Value used during Indexing")

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
