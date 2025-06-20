# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Build and Install
```bash
make build    # Build the application to output/godevmcp
make install  # Install to $GOPATH/bin/godevmcp
```

### Testing and Quality
```bash
make test     # Run all unit tests with verbose output
make fmt      # Format code using gofumpt and golangci-lint
make lint     # Run golangci-lint for code quality checks
```

### Development
```bash
make run      # Run the application in serve mode
make inspect  # Run with MCP inspector for debugging
```

## Architecture Overview

GoDevMCP is a Model Context Protocol (MCP) server that provides Go development tools. The architecture follows clean architecture principles:

### Core Components
- **main.go**: Entry point using Google's subcommands pattern
- **internal/subcmd/**: Subcommand implementations (serve, tree, datetime, etc.)
- **internal/mcptool/**: MCP tool handlers and registration
- **internal/app/**: Core business logic for each tool
- **internal/infra/**: Infrastructure services (filesystem, GitHub API, HTTP client)
- **internal/repository/**: Data access layer implementations

### MCP Tools Architecture
The server registers multiple tools in `internal/mcptool/register.go`:
- `get_current_datetime`: Current date/time
- `run_make`: Execute make targets with intelligent output filtering
- `tree_dir`: Directory structure visualization
- `search_godoc`/`read_godoc`: Go documentation tools
- `search_github_code`/`get_github_content`/`tree_github_repo`: GitHub integration
- `search_local_files`: Local file content search

### Key Patterns
- Each tool has a handler in `internal/mcptool/` that delegates to business logic in `internal/app/`
- Infrastructure concerns are abstracted through interfaces in `internal/repository/`
- Subcommands provide CLI access to tools for testing/debugging
- The server supports both stdio and SSE modes for different MCP clients

### Testing Strategy
- Unit tests are colocated with source files (e.g., `github_test.go`, `tree_test.go`)
- Test MCP tools with `make test-mcp-tool`
- Use testify for assertions and mocking

## VSCode Integration
Add to your `settings.json`:
```json
"mcp": {
    "servers": {
        "go-dev-mcp": {
            "type": "stdio",
            "command": "godevmcp",
            "args": ["serve"]
        }
    }
}
```