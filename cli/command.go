package cli

import (
	"fmt"
	"gemini_cli_tool/fileinfo"

	"github.com/spf13/cobra"
)

var opts chatOpts

func NewConfigCommand() *cobra.Command {
	var addDirectories []string
	var deleteDirectories []string
	var addSkipTypes []string
	var deleteSkipTypes []string
	var addSkipFiles []string
	var deleteSkipFiles []string
	var relevanceIndex float32
	var showConfig bool
	var addAPIKeys []string
	var deleteAPIKeys []string
	var fileEdit bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Set configurations for indexing",

		RunE: func(cmd *cobra.Command, args []string) error {

			if showConfig {
				config, err := LoadConfig()
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				showConfigFormatted(config)
				return nil
			}

			if fileEdit {
				fmt.Println(fileinfo.Yellow("Opening configuration file for editing..."))
				EditConfig()
				return nil
			}

			return setConfig(addDirectories, deleteDirectories, addSkipTypes, deleteSkipTypes, addSkipFiles, deleteSkipFiles, addAPIKeys, deleteAPIKeys, relevanceIndex)
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
	cmd.Flags().StringSliceVar(&addAPIKeys, "add-apikeys", []string{}, "List of API keys to add")
	cmd.Flags().StringSliceVar(&deleteAPIKeys, "del-apikeys", []string{}, "List of API keys to remove")
	cmd.Flags().BoolVarP(&fileEdit, "edit", "e", false, "Open the configuration file in an editor")

	return cmd
}

func NewIndexCommand(hs *fileinfo.HashSet) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Index files in the configured directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return indexFilesCmd(hs)
		},
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

func NewChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Gives you a Chatting platform with alot of functionalities",
		Long: `
-----------------------------------------------------------------------------------------------------------------
Gives you a Chatting platform with alot of functionalities
		
Functionalities:
	- Command Prefix       : "$"                   (All system commands start with this prefix)
	- Quit Command         : "$quit"               (Exits the CLI application)
	- Purge History        : "$purge"              (Clears the command history)
	- Toggle Input Mode    : "$mode"               (Switches between single-line and multi-line input modes)
	- Toggle Output Format : "$format"             (Enables or disables formatted output)
	- Set Style            : "$style <style_name>" (Sets the output style.)
	- Index Files          : "$index"              (Indexes files in the specified directory for search purposes)
	- Search Files         : "$search <query>"     (Searches indexed files based on the provided query)
 
Different styles include:
	- AsciiStyle           : "ascii"               (ASCII-art inspired style)
	- AutoStyle            : "auto"                (Automatically selects the best style based on the terminal)
	- DarkStyle            : "dark"                (Dark theme with high contrast)
	- DraculaStyle	       : "dracula"             (Inspired by the Dracula color scheme)
	- TokyoNightStyle      : "tokyo-night"         (Inspired by the Tokyo Night color scheme)
	- LightStyle	       : "light"               (Light theme for bright environments)
	- NoTTYStyle	       : "notty"               (No TTY output style, minimal formatting)
	- PinkStyle            : "pink"                (A playful pink color scheme)
-----------------------------------------------------------------------------------------------------------------`,
		RunE: startchat,
	}

	cmd.Flags().BoolVarP(&opts.Format, "format", "f", true, "render markdown-formatted response")
	cmd.Flags().StringVarP(&opts.Style, "style", "s", "auto", "markdown format style (ascii, dark, light, pink, notty, dracula)")
	cmd.Flags().BoolVarP(&opts.Multiline, "multiline", "m", false, "read input as a multi-line string")
	cmd.Flags().StringVarP(&opts.Terminator, "term", "t", "~", "multi-line input terminator")

	return cmd
}
