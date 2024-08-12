# gemini_powered_cli_tool
A gemini powered interactable CLI tool that keeps record of all your files and categorizes them based on their metadata. Utilizes latest capabilities of Gemini-1.5-flash to generate ,store and search for most appropiate file based on your query.

## Requirements

- Go programming language (version 1.22.5)
- Gemini API key (sign up [here](https://geminiapi.com))

## Installation

1. Clone the repository :
   ```bash
   git clone https://github.com/lakshyajain-0291/gemini_powered_cli_tool.git
   cd gemini_powered_cli_tool
   ```

2. Build the CLI tool :
   ```bash
   go build -o gencli
   ```

3. Set up your Gemini API Keys :
   ```bash
   ./gencli config --add-apikeys "<your-api-key-1>","<your-api-key-1>"
   ```

4. Set up directories to Index :
   ```bash
   ./gencli config --add-dir "<directory-path-1>","<directory-path-2>"
   ```

## Usage

### Starting the CLI Chat

```bash
./gencli chat
```

#### Chat Commands

- **Quit Application**    : `$quit` -  Exit the CLI tool. 
- **Purge chat History**  : `$purge` - Clear the chat history.
- **Toggle Input Mode**: `$mode` - Toggle between single-line and multi-line input modes.
- **Toggle Format**       : `$format` - Toggle response formatting (Markdown).
- **Set Style**: `$style <style>` - Set the formatting style (ascii, dark, light, pink, notty, dracula, tokyo-night).
- **Index**: `$index` - Indexes configured files in chat interface.
- **Search**: `$search <query>` - Searchs indexed files based on the query provided.


### Starting the CLI Index

```bash
./gencli index
```

### Searching

```bash
./gencli search "<query>"
```

### CLI -help
```bash

gencli -h

-----------------------------------------------------------------------------------------------------------------
Welcome to the Gemini-based interactive CLI tool!

This versatile command-line interface is designed to enhance your file management experience with :

        - Chat   : Engage in intelligent conversations with your ternimal.
        - Index  : Efficiently index your files for faster search and retrieval.
        - Search : Perform detailed searches to quickly find the files you need.
-----------------------------------------------------------------------------------------------------------------

Usage:
  gencli [flags]
  gencli [command]

Available Commands:
  chat        Gives you a Chatting platform with alot of functionalities
  completion  Generate the autocompletion script for the specified shell
  config      Set configurations for the indexing
  help        Help about any command
  index       Index files in the configured directories
  search      Search files based on the provided query.

Flags:
  -h, --help      help for gencli
  -v, --version   version for gencli

Use "gencli [command] --help" for more information about a command.
```

### CLI chat -help
```bash

gencli chat -h

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
        - DraculaStyle         : "dracula"             (Inspired by the Dracula color scheme)
        - TokyoNightStyle      : "tokyo-night"         (Inspired by the Tokyo Night color scheme)
        - LightStyle           : "light"               (Light theme for bright environments)
        - NoTTYStyle           : "notty"               (No TTY output style, minimal formatting)
        - PinkStyle            : "pink"                (A playful pink color scheme)
-----------------------------------------------------------------------------------------------------------------

Usage:
  gencli chat [flags]

Flags:
  -f, --format         render markdown-formatted response (default true)
  -h, --help           help for chat
  -m, --multiline      read input as a multi-line string
  -s, --style string   markdown format style (ascii, dark, light, pink, notty, dracula) (default "auto")
  -t, --term string    multi-line input terminator (default "~")
```

### CLI config -help
```bash

gencli config -h
Set configurations for the indexing

Usage:
  gencli config [flags]

Flags:
      --add-apikeys strings     List of API keys to add
      --add-dir strings         List of directories to add to the index
      --add-skipfiles strings   List of files to skip during indexing
      --add-skiptypes strings   List of file types to skip during indexing
      --del-apikeys strings     List of API keys to remove
      --del-dir strings         List of directories to remove from the index
      --del-skipfiles strings   List of files to stop skipping during indexing
      --del-skiptypes strings   List of file types to stop skipping during indexing
  -h, --help                    help for config
  -r, --relindex float32        Relevance Value used during Indexing (default 0.8)
  -s, --show-config             Show the current configuration
```

## Contributing

Contributions are welcome! Fork the repository and submit a pull request.

## License


---