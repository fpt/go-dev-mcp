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

vscode-build: ## Build binaries for VSCode extension
	mkdir -p vscext/server
	GOOS=windows GOARCH=amd64 go build -o vscext/server/godevmcp-win64.exe godevmcp/main.go
	GOOS=windows GOARCH=arm64 go build -o vscext/server/godevmcp-win-arm64.exe godevmcp/main.go
	GOOS=darwin GOARCH=arm64 go build -o vscext/server/godevmcp-darwin-arm64 godevmcp/main.go
	GOOS=linux GOARCH=386 go build -o vscext/server/godevmcp-linux-x86 godevmcp/main.go
	GOOS=linux GOARCH=amd64 go build -o vscext/server/godevmcp-linux-amd64 godevmcp/main.go

vscode-check: ## Check TypeScript code for extension
	cd vscext && npm run compile

vscode-package: vscode-build vscode-check ## Build and package VSCode extension
	cd vscext && vsce package

vscode-publish: vscode-build vscode-check ## Build and publish VSCode extension
	cd vscext && vsce publish

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'
