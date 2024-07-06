# gemini_powered_cli_tool
A gemini powered interactable CLI tool that keeps record of all your files and categorizes them based on their metadata.

## Features

- **Chat Interface**: Communicate interactively with Gemini API.
- **System Scanning**: Scan and index system files and metadata.
- **Command Line Interface**: Easy-to-use command line interface for interacting with the tool.

## Requirements

- Go programming language (version 1.22.5)
- Gemini API key (sign up [here](https://geminiapi.com))

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your/repository.git
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
./gencli chat
```

### Commands

- **Quit Application**: `$q` -  Exit the CLI tool. 
- **Purge Chat Hisroty**: `$p` - Clear the chat history.
- **Toggle Input Mode**: `$m` - Toggle between single-line and multi-line input modes.
- **Toggle Format**: `$tf` - Toggle response formatting (Markdown).
- **Set Style**: `$st <style>` - Set the formatting style (ascii, dark, light, pink, notty, dracula).

### Example Session

```
$ ./gencli chat

LAPTOP-P9DU6BMB\LAKSHYA JAIN> Hello
gemini>
  Hello there! How can I assist you today?

  I am Gemini, a multi-modal AI model, developed by Google. I have been trained
  on a massive amount of text data, and I am capable of understanding and
  generating human language, answering questions, and providing information on
  a wide range of topics.

  Is there anything I can help you with today?

LAPTOP-P9DU6BMB\LAKSHYA JAIN> $tf
cli>                          Toggled format to true.

LAPTOP-P9DU6BMB\LAKSHYA JAIN> $st dark
cli>                          Set style to dark.

LAPTOP-P9DU6BMB\LAKSHYA JAIN> How are you?
gemini>
  I am well, thank you for asking. As an AI, I do not have personal feelings
  or emotions, but I am designed to be helpful and informative.

  Is there anything I can assist you with today? I can answer your questions,
  provide information on various topics, or help you with tasks such as
  writing emails or translating languages.

LAPTOP-P9DU6BMB\LAKSHYA JAIN> $m
cli>                          Switched to single-line input mode.

LAPTOP-P9DU6BMB\LAKSHYA JAIN> $q
cli>                          Exiting gen-cli..

```

## Contributing

Contributions are welcome! Fork the repository and submit a pull request.

## License


---

Replace placeholders such as `your_api_key_here`, `X.X.X` with actual values and versions relevant to your project. Customize the commands, features, and examples to fit the specifics of your Gemini-powered CLI tool.