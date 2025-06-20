# GoDevMCP

## Introduction

GoDevMCP provides convenient tools for Go development with Model Context Protocol (MCP) integration.

## Installation

### Prerequisites

- GitHub `gh` command is required.

### Using go install

You can install GoDevMCP directly using Go's install command:

```bash
go install fujlog.net/godev-mcp/godevmcp@latest
```

This will download, compile, and install the binary to your `$GOPATH/bin` directory (typically `~/go/bin`). Make sure this directory is in your system's PATH.

### Building from Source

1. Clone the repository
2. Run `make build` to build the application and `make install` to install.

### VSCode

Add this section in your user's `settings.json`
```
    "mcp": {
        "servers": {
            "go-dev-mcp": {
                "type": "stdio",
                "command": "godevmcp",
                "args": [
                    "serve"
                ],
            }
        }
    }
```

## Tools

### run_make

Runs make command for common development tasks.

### tree_dir

Returns directory tree structure for project navigation.

### search_godoc

Searches for Go packages on pkg.go.dev and returns matching results.

**Usage:**
```
search_godoc query
```

**Example:**
```
search_godoc html
```

This will search for packages related to "html" and return a list of matching packages with their descriptions.

### read_godoc

Fetches and displays documentation for a specific Go package.

**Usage:**
```
read_godoc package_url
```

**Example:**
```
read_godoc golang.org/x/net/html
```

This will retrieve the documentation for the specified package, including descriptions, functions, types, and examples.

### extract_function_names

Extracts exported function names from Go source files in a directory.

**Usage:**
```
extract_function_names directory
```

**Example:**
```
extract_function_names ./internal/app
```

This will recursively scan the directory for `.go` files (excluding test files) and extract all exported function names, showing both regular functions and methods with their receiver types.

### extract_call_graph

Analyzes function call relationships within a single Go file.

**Usage:**
```
extract_call_graph file_path
```

**Example:**
```
extract_call_graph ./internal/app/github.go
```

This will show which functions call which other functions, including external package calls and local function calls. Useful for understanding code dependencies and refactoring impact analysis.

## Instructions

1. **Build the application**
   ```
   make build
   ```

2. **Run the application**
   ```
   make run
   ```

## Available Make Commands

- `make run` - Run the application
- `make build` - Build the application and install to ~/bin
- `make test` - Run unit tests
- `make fmt` - Format code using gofumpt
- `make lint` - Run golangci-lint
- `make inspect` - Run in MCP inspector
- `make help` - Display help information

## Development

### Project Structure

```
.
├── godevmcp/       # Application entry point
│   └── main.go     # Main application entry point
├── doc/            # Documentation files
├── internal/       # Private application and library code
│   ├── app/        # Application core functionality
│   ├── infra/      # Infrastructure code
│   ├── mcptool/    # MCP tooling implementations
│   ├── repository/ # Repository implementations
│   └── subcmd/     # Subcommand implementations
├── output/         # Build artifacts
│   └── godevmcp    # Compiled binary
├── pkg/            # Public library code
│   └── htmlu/      # HTML utility package
├── Makefile        # Build automation
├── go.mod          # Go module definition
└── go.sum          # Go module checksum
```

### Common Development Workflow

1. Make changes to the code
2. Run `make fmt` to format code
3. Run `make lint` to check for issues
4. Run `make test` to ensure tests pass
5. Build with `make build`

### For AI Development

- Follow instructions in CONTRIBUTING.md
- Use `run_make` and `tree_dir` tools rather than using shell commands.
- Use `search_godoc` and `read_godoc` tools to understand how to use depending packages.
- Use `tree_github_repo`, `search_github_code`,  `get_github_content` to inspect github repository.
- Remember to update README.md when making significant changes
