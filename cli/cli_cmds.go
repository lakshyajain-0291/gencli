package cli

import (
	"bufio"
	"errors"
	"fmt"
	"gemini_cli_tool/fileinfo"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
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
	setup *setup
}

var _ command = (*systemCommand)(nil)

func newSystemCommand(setup *setup) command {
	return &systemCommand{
		setup: setup,
	}
}

// run implements command.
func (c *systemCommand) run(message string) bool {
	switch message {
	case systemCmdQuit:
		c.print("Exiting gen-cli..")
		return true

	case systemCmdPurgeHistory:
		c.setup.session.ClearHistory()
		c.print("Cleared the setup history.")

	case systemCmdToggleInputMode:
		if c.setup.opts.Multiline {
			c.setup.reader.HistoryEnable()
			c.setup.opts.Multiline = false
			c.print("Switched to single-line input mode.")
		} else {
			c.setup.reader.HistoryDisable()
			c.setup.opts.Multiline = true
			c.print("Switched to Multi-line input mode.")

		}

	case systemCmdToggleFormat:
		c.setup.opts.Format = !c.setup.opts.Format
		c.print(fmt.Sprintf("Toggled format to %v.", c.setup.opts.Format))

	default:
		if strings.HasPrefix(message, systemCmdSetStyle) {
			style := strings.TrimPrefix(message, systemCmdSetStyle+" ")
			c.setup.opts.Style = style
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
				c.print("Most relevelent file is : ")
				c.print(fmt.Sprintf("\nFile : %s\nDirectory : %s\nDescription : %s\n", file.Name, file.Directory, file.Description))
			}

		} else {
			c.print("Unknown system command.")
		}
	}

	return false
}

func (c *systemCommand) print(message string) {
	fmt.Printf("%s%s\n\n", c.setup.prompt.cli, message)
}

type geminiCommand struct {
	setup   *setup
	spinner *fileinfo.Spinner
	writer  *bufio.Writer
}

var _ command = (*geminiCommand)(nil)

func newGeminiCommand(setup *setup) command {
	writer := bufio.NewWriter(os.Stdout)
	return &geminiCommand{
		setup:   setup,
		spinner: fileinfo.NewSpinner(5, time.Second, writer),
		writer:  writer,
	}
}

// run implements command.
func (g *geminiCommand) run(message string) bool {
	g.printFlush(g.setup.prompt.gemini)
	g.spinner.Start()

	if g.setup.opts.Format {
		g.runBlocking(message)
	} else {
		g.runStreaming(message)
	}

	return false
}

func (g *geminiCommand) runBlocking(message string) {
	// fmt.Println("Giving formatted output")

	response, err := g.setup.session.SendMessage(message)
	g.spinner.Stop()

	if err != nil {
		fmt.Print(Red(err.Error()))
	} else {
		var builder strings.Builder

		for _, candidate := range response.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}

		output, err := glamour.Render(builder.String(), g.setup.opts.Style)
		if err != nil {
			fmt.Printf(Red("Failed to format : %s\n"), err)
			g.setup.opts.Format = false
			fmt.Println((builder.String()))
			return
		}

		fmt.Print(output)
	}
}

func (g *geminiCommand) runStreaming(message string) {
	// fmt.Println("Giving streaming output")

	responseIterator := g.setup.session.SendMessageStream(message)
	g.spinner.Stop()
	for {
		response, err := responseIterator.Next()
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				fmt.Print(Red(err.Error()))
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
