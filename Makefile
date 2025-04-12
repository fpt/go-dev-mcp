run: ## Run the application
	go run cmd/main.go serve

build: ## Build the application
	go build -o output/godevmcp cmd/main.go
	cp output/godevmcp ~/bin/godevmcp
	chmod +x ~/bin/godevmcp

test: ## Run unit tests
	go test -v ./...

fmt: ## Run format
	gofumpt -extra -w .

lint: ## Run lint
	golangci-lint run

inspect: ## Run in MCP inspector
	npx @modelcontextprotocol/inspector go run ./cmd/main.go serve

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "%-20s %s\n", $$1, $$2}'
