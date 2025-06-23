run: ## Run the application
	go run godevmcp/main.go serve

build: ## Build the application
	go build -o output/godevmcp godevmcp/main.go

install: ## Install the application
	go install ./godevmcp
	@echo "Installed to $(shell go env GOPATH)/bin/godevmcp"

test: ## Run unit tests
	go test -v ./...

test-mcp-tool: ## Run MCP tool tests
	(echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}'; echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "search_godoc", "arguments": {"query": "mcp-go"}}}') | ./output/godevmcp serve

fmt: ## Run format
	gofumpt -extra -w .
	golangci-lint fmt

lint: ## Run lint
	golangci-lint run

fix: ## Run fix
	golangci-lint run --fix

inspect: ## Run in MCP inspector
	npx @modelcontextprotocol/inspector go run ./godevmcp/main.go serve

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'
