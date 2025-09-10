package cli

import (
	"bufio"
	"errors"
	"fmt"
	"gemini_cli_tool/fileinfo"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	// "github.com/charmbracelet/glamour/ansi"

	"google.golang.org/api/iterator"
)

const (
	systemCmdPrefix          = "$"
	systemCmdQuit            = "$quit"
	systemCmdPurgeHistory    = "$purge"
	systemCmdToggleInputMode = "$mode"
	systemCmdToggleFormat    = "$format"
	systemCmdSetStyle        = "$style"
	systemCmdIndex           = "$index"
	systemCmdSearch          = "$search"
)

type command interface {
	run(message string) bool
}

type systemCommand struct {
	chat *chat
}

var _ command = (*systemCommand)(nil)

func newSystemCommand(chat *chat) command {
	return &systemCommand{
		chat: chat,
	}
}

// run implements command.
func (c *systemCommand) run(message string) bool {
	switch message {
	case systemCmdQuit:
		c.print("Exiting gen-cli..")
		return true

	case systemCmdPurgeHistory:
		c.chat.session.ClearHistory()
		c.print("Cleared the chat history.")

	case systemCmdToggleInputMode:
		if c.chat.opts.Multiline {
			c.chat.reader.HistoryEnable()
			c.chat.opts.Multiline = false
			c.print("Switched to single-line input mode.")
		} else {
			c.chat.reader.HistoryDisable()
			c.chat.opts.Multiline = true
			c.print("Switched to Multi-line input mode.")

		}

	case systemCmdToggleFormat:
		c.chat.opts.Format = !c.chat.opts.Format
		c.print(fmt.Sprintf("Toggled format to %v.", c.chat.opts.Format))

	default:
		if strings.HasPrefix(message, systemCmdSetStyle) {
			style := strings.TrimPrefix(message, systemCmdSetStyle+" ")
			c.chat.opts.Style = style
			c.print(fmt.Sprintf("Set style to %s.", style))

		} else if strings.HasPrefix(message, systemCmdIndex) {
			hashSet := fileinfo.NewHashSet() // Initialize the hash set
			// Load the hash set from file at the start
			if err := hashSet.LoadFromFile(); err != nil {
				return true
			}

			err := indexFiles(hashSet)
			// spinners.stop()
			if err != nil {
				c.print(err.Error())
			} else {
				c.print("Indexing completed successfully.")
			}

		} else if strings.HasPrefix(message, systemCmdSearch) {
			query := strings.TrimPrefix(message, systemCmdSetStyle+" ")
			file, err := searchFiles(query)
			// spinners.stop()

			if err != nil {
				c.print(err.Error())
			} else {
				c.print("Most relevant file is : ")
				c.print(fmt.Sprintf("\nFile : %s\nDirectory : %s\nDescription : %s\n", file.Name, file.Directory, file.Description))
				c.print(("Do you want to open this file? (y/n): "))

				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "y" {
					err = OpenFileWithDefaultApp(file.Directory)
					if err != nil {
						c.print(fmt.Sprintf("Failed to open file: %v", err))
					}
				}
			}

		} else {
			c.print("Unknown system command.")
		}
	}

	return false
}

func (c *systemCommand) print(message string) {
	fmt.Printf("%s%s\n\n", c.chat.prompt.Cli, message)
}

// OpenFileWithDefaultApp opens the file with the default application based on OS.
func OpenFileWithDefaultApp(path string) error {
	var cmd string
	var args []string

	switch {
	case strings.Contains(strings.ToLower(os.Getenv("OS")), "windows"):
		cmd = "cmd"
		args = []string{"/c", "start", "", path}
	case strings.Contains(strings.ToLower(os.Getenv("XDG_SESSION_TYPE")), "wayland") ||
		strings.Contains(strings.ToLower(os.Getenv("XDG_SESSION_TYPE")), "x11") ||
		os.Getenv("DISPLAY") != "":
		cmd = "xdg-open"
		args = []string{path}
	case os.Getenv("TERM_PROGRAM") == "Apple_Terminal" || os.Getenv("OSTYPE") == "darwin":
		cmd = "open"
		args = []string{path}
	default:
		// Try xdg-open as a fallback
		cmd = "xdg-open"
		args = []string{path}
	}

	return execCommand(cmd, args...)
}

func execCommand(name string, arg ...string) error {
	// fmt.Printf("Executing command: %s %s\n", name, strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)
	return cmd.Start()
}

type geminiCommand struct {
	chat    *chat
	spinner *fileinfo.Spinner
	writer  *bufio.Writer
}

var _ command = (*geminiCommand)(nil)

func newGeminiCommand(chat *chat) command {
	writer := bufio.NewWriter(os.Stdout)
	return &geminiCommand{
		chat:    chat,
		spinner: fileinfo.NewSpinner(5, time.Second, writer),
		writer:  writer,
	}
}

// run implements command.
func (g *geminiCommand) run(message string) bool {
	g.printFlush(g.chat.prompt.Gemini)
	g.spinner.Start()

	if g.chat.opts.Format {
		g.runBlocking(message)
	} else {
		g.runStreaming(message)
	}

	return false
}

func (g *geminiCommand) runBlocking(message string) {
	// fmt.Println("Giving formatted output")

	response, err := g.chat.session.SendMessage(message)
	g.spinner.Stop()
	// fmt.Println("Spinner stopped")
	if err != nil {
		fmt.Print(fileinfo.Red(err.Error()))
	} else {
		var builder strings.Builder
		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}

		// Get the selected style
		output, err := glamour.Render(builder.String(), g.chat.opts.Style)
		if err != nil {
			fmt.Printf(fileinfo.Red("Failed to format : %s\n"), err)
			g.chat.opts.Format = false
			fmt.Println((builder.String()))
			return
		}
		fmt.Print(output)
	}

	// // Check if chat is nil
	// if g.chat == nil {
	// 	fmt.Println("Chat is nil. Make sure it is properly initialized.")
	// 	return
	// }

	// // Check if session is nil
	// if g.chat.session == nil {
	// 	fmt.Println("Session is nil. Make sure it is properly initialized.")
	// 	return
	// }

	// // Print the session details for debugging
	// fmt.Printf("Session details: %+v\n", g.chat.session)

	// if err != nil {
	// 	// If the error is from SendMessage, print it out
	// 	fmt.Println("Error sending message to Gemini API:", err)

	// 	// Add more context to the error if possible
	// 	if strings.Contains(err.Error(), "404") {
	// 		fmt.Println("The endpoint was not found. Check the API URL and endpoint path.")
	// 	} else if strings.Contains(err.Error(), "403") {
	// 		fmt.Println("Authentication failed. Check your API key or token.")
	// 	} else if strings.Contains(err.Error(), "500") {
	// 		fmt.Println("Server error. The Gemini API might be down.")
	// 	} else {
	// 		fmt.Println("An unknown error occurred.")
	// 	}

	// 	return
	// }
}

func (g *geminiCommand) runStreaming(message string) {
	// fmt.Println("Giving streaming output")

	responseIterator := g.chat.session.SendMessageStream(message)
	g.spinner.Stop()
	for {
		response, err := responseIterator.Next()
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				fmt.Print(fileinfo.Red(err.Error()))
			}
			break
		}
		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				g.printFlush(fmt.Sprintf("%s", part))
			}
		}
	}

	fmt.Print("\n\n")
}

func (g *geminiCommand) printFlush(message string) {
	fmt.Fprintf(g.writer, "%s", message)
	g.writer.Flush()
}
