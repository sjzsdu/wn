# WN - Multi-functional Command Line Tool

[中文](README.md) | English

## Overview
WN is a multi-functional command line tool that provides practical features such as file packaging and code statistics to help developers manage and analyze code more efficiently.

WN integrates various utilities and AI capabilities, aiming to enhance developers' daily work efficiency. It not only provides powerful file packaging functionality, supporting multiple output formats and intelligent file processing, but also includes detailed code statistics analysis tools to help developers better understand and manage their codebase. By integrating multiple large language models (such as OpenAI and DeepSeek), WN also offers intelligent AI conversation capabilities, coupled with a customizable Agent system that can provide professional technical support and advice for different scenarios. Whether it's code management, project analysis, or technical consultation, WN can serve as a powerful assistant for developers, providing comprehensive support.

## Installation

### Quick Install
```bash
curl -sSL https://raw.githubusercontent.com/sjzsdu/wn/refs/heads/master/get-wn.sh | bash
```

### Build from Source
1. Ensure Go 1.20+ is installed
2. Clone the repository
3. Run `go build` to compile the project

## Main Features

### 1. File Packaging (pack)
Package source code files of specified types into various document formats. (For knowledge base training)

#### Features
- Multiple Output Formats
  - PDF (Optimized font rendering, Chinese character support)
  - Markdown
  - XML
  - Plain text
- Intelligent File Processing
  - Supports all text file formats
  - Automatic readable text file detection
  - File exclusion patterns
  - .gitignore rules support
- Git Repository Support
  - Direct cloning and packaging of Git repositories
- Multi-language Support
  - Chinese and English interface switching

#### Usage
```bash
wn pack [flags]
```

##### Parameters
- Global Parameters
  - `-p, --workPath`: Specify working directory (default: current directory)
- Packaging Parameters
  - `-e, --exts`: File extensions to include (default: *, means all files)
  - `-o, --output`: Output filename (default: output.xml)
  - `-x, --excludes`: File patterns to exclude
  - `-g, --git-url`: Git repository URL for direct cloning and packaging
  - `-d, --disable-gitignore`: Disable .gitignore rule processing

#### Examples
1. Package all files to PDF:
```bash
wn pack -o output.pdf
```

2. Package specific extensions to Markdown, excluding test directory:
```bash
wn pack -e go,md -o output.md -x "test/*"
```

3. Clone and package from Git repository:
```bash
wn pack -g https://github.com/sjzsdu/EventTrader.git -o trader-code.pdf
```

### 2. Code Statistics (static)
Analyze various metrics of project code to help developers understand code structure and quality.

#### Features
- Code Volume Statistics
  - Total line count
  - Code line count
  - Comment line count
  - Blank line count
- File Analysis
  - Statistics by language type
  - File count statistics
  - File size statistics
- Intelligent Recognition
  - Automatic programming language detection
  - .gitignore rules support

#### Usage
```bash
wn static [flags]
```

##### Parameters
- `-p, --path`: Specify statistics directory (default: current directory)
- `-e, --exts`: Specify file extensions to analyze
- `-x, --excludes`: File patterns to exclude
- `-d, --detail`: Show detailed statistics

#### Examples
1. Analyze current directory:
```bash
wn static
```

2. Analyze Go files in specified directory:
```bash
wn static -p /path/to/project -e go
```

3. Show detailed statistics:
```bash
wn static -d
```

## Configuration

### Global Configuration
Use `wn config` command to manage global settings:

```bash
wn config [flags]
```

#### Configuration Items
- `--lang`: Set interface language (default: en)
  - Chinese interface: `wn config --lang zh`
  - English interface: `wn config --lang en`
- `--default_provider`: Set default LLM provider
- `--default_agent`: Set default Agent
- `--deepseek_apikey`: Set DeepSeek API key
- `--deepseek_model`: Set DeepSeek default model
- `--openai_apikey`: Set OpenAI API key
- `--openai_model`: Set OpenAI default model
- `--list`: List all current configurations

### 3. AI Conversation (ai)
Intelligent conversation with AI assistants, supporting multiple large language models.

#### Features
- Multiple Model Support
  - OpenAI
  - DeepSeek
- Stream Output
- Custom Agent Support
- Context Memory

#### Usage
```bash
wn ai [flags]
```

##### Parameters
- `-c, --provider`: Specify LLM provider
- `-m, --model`: Specify model to use
- `-t, --max-tokens`: Maximum response tokens (default: 2000)
- `-a, --agent`: Specify Agent to use
- `--providers`: List available LLM providers
- `--models`: List available models for current provider

### 4. Agent Management
Agents are preset AI roles that can help complete specific tasks.

#### Usage
```bash
wn agent [flags]
```

##### Parameters
- `--list`: List all available Agents
- `--create <name>`: Create or update Agent
- `--delete <name>`: Delete specified Agent
- `--show <name>`: Show Agent content
- `--content <text>`: Set Agent content
- `--file <path>`: Read Agent content from file

## Future Plans
- File difference comparison
- Project documentation generation
- Code quality checking
- More features in development...

## Contributing
PRs and issues are welcome to help improve this tool.

## License
MIT