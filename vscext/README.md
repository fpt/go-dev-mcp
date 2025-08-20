# Go Development MCP Server

A Model Context Protocol (MCP) server that provides comprehensive Go development tools for VS Code's AI features.

## Features

This extension provides an MCP server with the following tools:

### Documentation & Package Management
- **Search Go Documentation** - Search pkg.go.dev for Go packages
- **Read Go Documentation** - Read detailed documentation for specific packages
- **Search within Documentation** - Find specific content within package docs

### GitHub Integration
- **Search GitHub Code** - Search for Go code across GitHub repositories
- **Get GitHub Content** - Retrieve file contents from GitHub repositories
- **GitHub Repository Tree** - Explore repository structure

### Local Development
- **Local File Search** - Search through local Go files by content
- **Directory Tree** - Display project directory structure

### Code Analysis & Validation
- **Go Code Validation** - Comprehensive static analysis using go vet, build checks, formatting validation, and module tidiness
- **AST Analysis** - Extract Go AST information and call graphs
- **Package Dependencies** - Analyze import dependencies in Go projects
- **Markdown Scanning** - Extract headings and structure from markdown files

## How to Use

1. Install this extension
2. The MCP server will automatically be available to VS Code's AI features
3. Use AI chat features - the server provides context about Go development

## Requirements

- VS Code with AI/Copilot features enabled
- Go development environment (for local file operations)

## Architecture

The server is implemented in Go and provides stdio-based MCP protocol communication. It includes:

- Clean architecture with separated concerns
- Comprehensive testing
- Cross-platform binary support
- Intelligent output filtering for better token efficiency

## Repository

Source code: [https://github.com/fpt/go-dev-mcp](https://github.com/fpt/go-dev-mcp)

## License

MIT License - see LICENSE file for details.