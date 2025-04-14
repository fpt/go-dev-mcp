# GoDevMCP

## Introduction

GoDevMCP provides convenient tools for Go development with Model Context Protocol (MCP) integration.

## Project Structure

```
.
├── cmd/            # Application entry points
│   └── main.go     # Main application entry point
├── doc/            # Documentation files
├── internal/       # Private application and library code
│   ├── app/        # Application core functionality
│   ├── infra/      # Infrastructure code
│   └── mcptool/    # MCP tooling implementations
├── output/         # Build artifacts
│   └── devmcp      # Compiled binary
├── pkg/            # Public library code
├── Makefile        # Build automation
├── go.mod          # Go module definition
└── go.sum          # Go module checksum
```

## Usage
### VSCode

Add this section in your user's `settings.json`
```
    "mcp": {
        "servers": {
            "go-dev-mcp": {
                "type": "stdio",
                "command": "/<path_to>/godevmcp",
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

## Getting Started

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
