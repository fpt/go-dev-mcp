# GoDevMCP

## Introduction

GoDevMCP provides convenient tools for Go development with Model Context Protocol (MCP) integration.

## Installation

### Prerequisites

- GitHub `gh` command is required.

### Using go install

You can install GoDevMCP directly using Go's install command:

```bash
go install github.com/fpt/go-dev-mcp/godevmcp@latest
```

This will download, compile, and install the binary to your `$GOPATH/bin` directory (typically `~/go/bin`). Make sure this directory is in your system's PATH.

### Building from Source

1. Clone the repository
2. Run `make build` to build the application and `make install` to install.

## MCP setup

### Claude code

Run this command to add to user scope
```
claude mcp add godevmcp -s user godevmcp serve
```

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

### search_local_files

Searches for text within files in a local directory with match limiting to reduce output size.

**Usage:**
```
search_local_files path query extension max_matches
```

**Example:**
```
search_local_files ./internal/app "SearchLocalFiles" go 10
```

This will search for the text "SearchLocalFiles" in all `.go` files within the `./internal/app` directory, showing up to 10 matches per file with line numbers.

### get_github_content

Retrieves the content of a specific file from a GitHub repository with optional line-based paging for large files.

**Usage:**
```
get_github_content repo path [offset] [limit]
```

**Parameters:**
- `repo`: GitHub repository in 'owner/repo' format
- `path`: Path to the file in the repository
- `offset`: Line number to start reading from (0-based, default: 0)
- `limit`: Number of lines to read (default: 100, 0 for all lines)

**Examples:**
```
# Get entire file
get_github_content owner/repo-name README.md

# Get first 20 lines
get_github_content owner/repo-name src/main.go 0 20

# Get lines 50-70 (20 lines starting from line 50)
get_github_content owner/repo-name src/main.go 50 20
```

This fetches content from GitHub files with optional pagination for large files, similar to the `read_godoc` tool.

### tree_github_repo

Displays the directory tree structure of a GitHub repository with depth limiting for efficient exploration.

**Usage:**
```
tree_github_repo repo path max_depth ignore_dot
```

**Example:**
```
tree_github_repo owner/repo-name "" 3 false
```

This will show the directory structure of the GitHub repository up to 3 levels deep.

### search_github_code

Searches for code patterns in GitHub repositories with compact formatting.

**Usage:**
```
search_github_code query language repo
```

**Example:**
```
search_github_code "func main" go owner/repo-name
```

This will search for "func main" in Go files within the specified repository, returning results in a compact format.

### get_current_datetime

Returns the current date and time.

**Usage:**
```
get_current_datetime
```

This tool requires no parameters and returns the current timestamp.

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
│   └── dq/         # Document query utility package
├── Makefile        # Build automation
├── go.mod          # Go module definition
└── go.sum          # Go module checksum
```

### Package Dependencies

The following Mermaid graph shows the internal package dependencies within the project:

```mermaid
graph TD
    %% Entry point
    main[godevmcp/main] --> subcmd[internal/subcmd]
    
    %% Core application layer
    subcmd --> app[internal/app]
    subcmd --> mcptool[internal/mcptool]
    mcptool --> app
    
    %% Application modules
    app --> repository[internal/repository]
    app --> infra[internal/infra]
    app --> contentsearch[internal/contentsearch]
    app --> model[internal/model]
    app --> dq[pkg/dq]
    
    %% MCP layer
    mcptool --> infra
    
    %% Infrastructure dependencies
    infra --> repository
    contentsearch --> model
    
    %% External dependencies (key ones)
    subcmd -.-> subcommands[google/subcommands]
    mcptool -.-> mcp[mark3labs/mcp-go]
    infra -.-> github[google/go-github]
    app -.-> errors[pkg/errors]
    app -.-> goldmark[yuin/goldmark]
    app -.-> cache[patrickmn/go-cache]
    
    %% Styling
    classDef entryPoint fill:#e1f5fe
    classDef coreLayer fill:#f3e5f5
    classDef appLayer fill:#e8f5e8
    classDef infraLayer fill:#fff3e0
    classDef external fill:#fce4ec,stroke-dasharray: 5 5
    
    class main entryPoint
    class subcmd,mcptool coreLayer
    class app,repository,infra,contentsearch,model appLayer
    class dq infraLayer
    class subcommands,mcp,github,errors,goldmark,cache external
```

This graph illustrates:
- **Entry Point**: `main` package coordinates subcommands
- **Core Layer**: `subcmd` (CLI) and `mcptool` (MCP server) provide user interfaces
- **Application Layer**: `app` contains business logic, supported by infrastructure packages
- **Infrastructure Layer**: Shared utilities and external integrations
- **External Dependencies**: Key third-party packages (shown with dashed lines)

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
