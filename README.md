# gemini_powered_cli_tool
A gemini powered interactable CLI tool that keeps record of all your files and categorizes them based on their metadata.

## Requirements

- Go programming language (version 1.22.5)
- Gemini API key (sign up [here](https://geminiapi.com))

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/lakshyajain-0291/gemini_powered_cli_tool.git
   cd gemini_powered_cli_tool
   ```

2. Set up your Gemini API key:
   ```bash
   export GEMINI_API_KEY=your_api_key_here
   ```

3. Build the CLI tool:
   ```bash
   go build -o gencli
   ```

## Usage

### Starting the CLI

```bash
./gencli setup
```

## Features

- **chat Interface**`./gencli setup`: Communicate interactively with Gemini API.

- **System Config**`./gencli config`: Config the directories,  filetypes and filenames to index.

- **System Indexing**`./gencli index`: Indexes configured files and their metadata.

- **System Searching**`./gencli search <query>`: Searches the indexed files based on the interactive query provided.

- **Command Line Interface**: Easy-to-use command line interface for interacting with the tool.

### Setup Commands

- **Quit Application**: `$quit` -  Exit the CLI tool. 
- **Purge setup Hisroty**: `$purge` - Clear the setup history.
- **Toggle Input Mode**: `$mode` - Toggle between single-line and multi-line input modes.
- **Toggle Format**: `$format` - Toggle response formatting (Markdown).
- **Set Style**: `$style <style>` - Set the formatting style (ascii, dark, light, pink, notty, dracula).
- **Index**: `$index` - Indexes configured files in chat interface.
- **Search**: `$search <query>` - Searchs indexed files based on the query provided.

### CLI -help
```bash

./gencli.exe -h   
Gemini Based interactable file manager CLI

Usage:
  gencli [flags]
  gencli [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Set configurations for the indexing
  help        Help about any command
  index       Index files in the configured directories
  search      Search files based on the provided query.
  setup       Starts a new setup session

Flags:
  -h, --help      help for gencli
  -v, --version   version for gencli

Use "gencli [command] --help" for more information about a command.

```

### CLI setup -help
```bash

.\gencli.exe setup -h
Starts a new setup session

Usage:
  gencli setup [flags]

Flags:
  -f, --format         render markdown-formatted response (default true)
  -h, --help           help for setup
  -m, --multiline      read input as a multi-line string
  -s, --style string   markdown format style (ascii, dark, light, pink, notty, dracula) (default "auto")
  -t, --term string    multi-line input terminator (default "~")

```

### CLI config -help
```bash

./gencli.exe config -h
Set configurations for the indexing

Usage:
  gencli config [flags]

Flags:
  -d, --directories strings     List of directories to index
  -h, --help                    help for config
  -r, --relativeindex float32   Relevance Value used during Indexing (default 0.8)        
  -f, --skipfiles strings       List of directories to index
  -t, --skiptypes strings       List of file types to skip during indexing

```

### CLI index -help
```bash

./gencli.exe index -h
Index files in the configured directories

Usage:
  gencli index [flags]

Flags:
  -h, --help   help for index

```


### CLI search -help
```bash

./gencli.exe search -h  
Search files based on the provided query.

Usage:
  gencli search [flags]

Flags:
  -a, --all    Display Name and Description of All Indexed files
  -h, --help   help for search

```
## Contributing

Contributions are welcome! Fork the repository and submit a pull request.

## License


---

Replace placeholders such as `your_api_key_here`, `X.X.X` with actual values and versions relevant to your project. Customize the commands, features, and examples to fit the specifics of your Gemini-powered CLI tool.