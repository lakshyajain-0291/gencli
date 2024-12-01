<div align="center">

# GenCLI

[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" width="60">](https://golang.org)
[<img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/googlecloud/googlecloud-original.svg" width="60">](https://cloud.google.com/ai)

[![Open in Visual Studio Code](https://img.shields.io/badge/Open%20in%20VS%20Code-007ACC?logo=visual-studio-code&logoColor=white)](https://vscode.dev/)
[![Contributors](https://img.shields.io/github/contributors/lakshyajain-0291/gemini_powered_cli_tool)](https://github.com/lakshyajain-0291/gemini_powered_cli_tool/graphs/contributors)
[![Forks](https://img.shields.io/github/forks/lakshyajain-0291/gemini_powered_cli_tool?style=social)](https://github.com/lakshyajain-0291/gemini_powered_cli_tool/network/members)
[![Stars](https://img.shields.io/github/stars/lakshyajain-0291/gemini_powered_cli_tool?style=social)](https://github.com/lakshyajain-0291/gemini_powered_cli_tool/stargazers)
[![License](https://img.shields.io/github/license/lakshyajain-0291/gemini_powered_cli_tool)](https://github.com/lakshyajain-0291/gemini_powered_cli_tool/blob/main/LICENSE)

*AI-Powered File Management and Intelligent Interaction CLI*

[Key Features](#key-features) • [Installation](#-installation) • [Usage](#-usage) • [Contributing](#-contributing)

</div>

## 🌟 Overview

GenCLI is a revolutionary CLI tool powered by Gemini-1.5-flash, designed to transform how you interact with and manage your files. By leveraging advanced AI capabilities, GenCLI provides intelligent file indexing, searching, and interactive chat functionalities.

## 🚀 Key Features

- 🤖 **AI-Powered File Management**: Intelligent file categorization and metadata analysis
- 🔍 **Advanced Search**: Quick file retrieval based on intelligent queries
- 💬 **Interactive CLI Chat**: Engage in conversations with your terminal
- ⚙️ **Flexible Configuration**: Easily customize indexing and search parameters
- 🎨 **Multiple Styling Options**: Customize your CLI experience with various themes

## 🌈 Why GenCLI?

- **Intelligent Indexing**: Automatically categorize and organize your files
- **Seamless Search**: Find files faster with AI-enhanced search capabilities
- **Customizable Experience**: Multiple themes and configuration options
- **Efficient Workflow**: Streamline file management and interaction
- **AI-Powered Insights**: Leverage Gemini's advanced language capabilities

## 📋 Prerequisites

- Go (version 1.22.5)
- Gemini API Key
- Git (optional)

## 🔧 Installation

<details>
<summary>Step-by-step guide</summary>

1. Clone the repository:
```bash
git clone https://github.com/lakshyajain-0291/gemini_powered_cli_tool.git
cd gemini_powered_cli_tool
```

2. Build the CLI tool:
```bash
go build -o gencli
```

3. Configure API Keys:
```bash
./gencli config --add-apikeys "YOUR_API_KEY_1","YOUR_API_KEY_2"
```

4. Configure Indexing Directories:
```bash
./gencli config --add-dir "/path/to/directory1","/path/to/directory2"
```
</details>

## 🎮 Available Commands

### Chat Mode
- `$quit`: Exit the application
- `$purge`: Clear chat history
- `$mode`: Toggle input mode
- `$format`: Toggle response formatting
- `$style`: Set formatting style
- `$index`: Index configured files
- `$search`: Search indexed files

### Styling Options
- ASCII
- Dark
- Light
- Pink
- Dracula
- Tokyo Night
- No TTY

## 🔍 Usage Examples

1. Start Chat Interface:
```bash
./gencli chat
```

2. Index Files:
```bash
./gencli index
```

3. Search Files:
```bash
./gencli search "your query"
```

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/NewAICapability`)
3. Commit your changes (`git commit -m 'Add advanced file categorization'`)
4. Push to the branch (`git push origin feature/NewAICapability`)
5. Open a Pull Request

## 📜 License

GenCLI is open-source software. Details available in the LICENSE file.

## 🙏 Acknowledgments

- Gemini API for powering intelligent interactions
- Go programming language community
- Open-source contributors

**Revolutionize your file management!** 📁🚀
